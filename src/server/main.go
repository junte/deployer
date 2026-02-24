package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

	args := extractArgs(request)
	componentName := request.FormValue("component")
	componentKey := request.FormValue("key")

	isAsync := request.Form["async"] != nil

	var deployErr error

	if isAsync {
		deployErr = deployAsync(componentName, componentKey, args)
	} else {
		deployErr = deploySync(componentName, componentKey, args, writer)
	}

	if deployErr != nil {
		http.Error(writer, fmt.Sprintf("deploy err: %v", deployErr), http.StatusBadRequest)
		return
	}
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
	componentName string,
	componentKey string,
	args map[string]string,
	writer http.ResponseWriter,
) error {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return errors.New("can't stream to response")
	}

	writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Connection", "keep-alive")

	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	output := make(chan string)
	finished := make(chan struct{})

	go func() {
		defer close(finished)

		for line := range output {
			err := writeSSEEvent(writer, flusher, "output", outputEventData{Message: line})
			if err != nil {
				log.WithError(err).Error("write output sse event")
			}
		}
	}()

	results, err := deployer.DeployComponent(
		&core.ComponentDeployRequest{
			ComponentName: componentName,
			ComponentKey:  componentKey,
			Args:          args,
			Output:        &output,
			IsAsync:       false,
		},
	)

	close(output)
	<-finished

	exitCode := 0
	if results != nil {
		exitCode = results.ExitCode
	}

	writeErr := writeSSEEvent(writer, flusher, "exit", exitEventData{ExitCode: exitCode})
	if writeErr != nil {
		log.WithError(writeErr).Error("write exit sse event")
	}

	if err != nil {
		return fmt.Errorf("deploy component: %w", err)
	}

	return nil
}

func deployAsync(
	componentName string,
	componentKey string,
	args map[string]string,
) error {
	_, err := deployer.DeployComponent(
		&core.ComponentDeployRequest{
			ComponentName: componentName,
			ComponentKey:  componentKey,
			Args:          args,
			Output:        nil,
			IsAsync:       true,
		},
	)

	return err
}

func extractArgs(request *http.Request) map[string]string {
	args := make(map[string]string)
	for key, values := range request.Form {
		args[key] = values[0]
	}

	return args
}
