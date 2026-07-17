package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"deployer/src/config"
	"deployer/src/core"
	"deployer/src/core/deployer"

	log "github.com/sirupsen/logrus"
)

func Run(configFile string) error {
	err := config.ReadConfig(configFile)
	if err != nil {
		return err
	}

	setupLogging()

	log.Infof("version: %s", config.Version)

	http.HandleFunc("/", handler)

	return startServer()
}

func setupLogging() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(
		&log.TextFormatter{
			DisableColors:          false,
			DisableLevelTruncation: false,
		},
	)
}

func startServer() error {
	server := &http.Server{
		Addr:              config.Config.Port,
		ReadHeaderTimeout: 5 * time.Second,
	}

	var err error

	if config.Config.TLS.Cert != "" && config.Config.TLS.Key != "" {
		log.Infof("starting https server on port %s", config.Config.Port)

		err = server.ListenAndServeTLS(config.Config.TLS.Cert, config.Config.TLS.Key)
	} else {
		log.Infof("starting http server on port %s", config.Config.Port)

		err = server.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func handler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "", http.StatusMethodNotAllowed)
		return
	}

	err := request.ParseForm()
	if err != nil {
		http.Error(writer, fmt.Sprintf("wrong query params err: %v", err), http.StatusBadRequest)
		return
	}

	isAsync := request.Form["async"] != nil

	deployRequest := &core.ComponentDeployRequest{
		ComponentName: request.FormValue("component"),
		ComponentKey:  request.FormValue("key"),
		Args:          extractArgs(request),
		IsAsync:       isAsync,
	}

	// resolve and validate the component before any response header is
	// written, so validation failures return a real HTTP 400 instead of
	// racing an already-committed 200 stream.
	componentDeployer, err := deployer.NewComponentDeployer(deployRequest)
	if err != nil {
		http.Error(writer, fmt.Sprintf("deploy err: %v", err), http.StatusBadRequest)
		return
	}

	if isAsync {
		go componentDeployer.DeployAsync()

		writer.WriteHeader(http.StatusOK)

		return
	}

	deploySync(componentDeployer, deployRequest, writer)
}

type outputEventData struct {
	Message string `json:"message"`
}

type exitEventData struct {
	ExitCode int `json:"exit_code"`
}

func writeSSEEvent(writer io.Writer, flusher http.Flusher, eventName string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal sse event %q: %w", eventName, err)
	}

	_, err = fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", eventName, payload)
	if err != nil {
		return fmt.Errorf("write sse event %q: %w", eventName, err)
	}

	flusher.Flush()

	return nil
}

func deploySync(
	componentDeployer *deployer.ComponentDeployer,
	request *core.ComponentDeployRequest,
	writer http.ResponseWriter,
) {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "can't stream to response", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Connection", "keep-alive")

	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	output := make(chan string)
	errorOutput := make(chan string)

	request.Output = &output
	request.ErrorOutput = &errorOutput

	var mu sync.Mutex

	var wg sync.WaitGroup

	wg.Go(func() {
		streamSSELines(writer, flusher, &mu, "output", output)
	})
	wg.Go(func() {
		streamSSELines(writer, flusher, &mu, "error", errorOutput)
	})

	results, err := componentDeployer.Deploy()

	close(output)
	close(errorOutput)
	wg.Wait()

	// the header is already committed as a 200 stream, so a deploy failure
	// can only be reported through SSE events, never a changed status code.
	if err != nil {
		log.WithError(err).Error("deploy component")

		writeErr := writeSSEEvent(writer, flusher, "error", outputEventData{Message: fmt.Sprintf("deploy err: %v\n", err)})
		if writeErr != nil {
			log.WithError(writeErr).Error("write error sse event")
		}
	}

	exitCode := resolveExitCode(results, err)

	writeErr := writeSSEEvent(writer, flusher, "exit", exitEventData{ExitCode: exitCode})
	if writeErr != nil {
		log.WithError(writeErr).Error("write exit sse event")
	}
}

func streamSSELines(
	writer io.Writer,
	flusher http.Flusher,
	mu *sync.Mutex,
	eventName string,
	lines <-chan string,
) {
	for line := range lines {
		mu.Lock()
		err := writeSSEEvent(writer, flusher, eventName, outputEventData{Message: line})
		mu.Unlock()

		if err != nil {
			log.WithError(err).Errorf("write %s sse event", eventName)
		}
	}
}

func resolveExitCode(results *core.ComponentDeployResults, err error) int {
	if results != nil {
		return results.ExitCode
	}

	if err != nil {
		return 1
	}

	return 0
}

func extractArgs(request *http.Request) map[string]string {
	args := make(map[string]string)
	for key, values := range request.Form {
		args[key] = values[0]
	}

	return args
}
