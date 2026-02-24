package server

import (
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

func Run() {
	config.ReadConfig()
	setupLogging()

	log.Infof("version: %s", config.Version)

	http.HandleFunc("/", handler)

	err := startServer()
	if err != nil {
		panic(err)
	}
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

	if err := request.ParseForm(); err != nil {
		http.Error(writer, fmt.Sprintf("wrong query params err: %v", err), http.StatusBadRequest)
		return
	}

	args := extractArgs(request)
	componentName := request.FormValue("component")
	componentKey := request.FormValue("key")

	var err error

	isAsync := request.Form["async"] != nil

	if isAsync {
		err = deployAsync(componentName, componentKey, args)
	} else {
		err = deploySync(componentName, componentKey, args, writer)
	}

	if err != nil {
		http.Error(writer, fmt.Sprintf("deploy err: %v", err), http.StatusBadRequest)
		return
	}
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
	defer close(output)

	done := make(chan int)

	go func() {
		for {
			select {
			case line, more := <-output:
				if !more {
					break
				}

				_, err := io.WriteString(writer, fmt.Sprintf("data: %s\n", line))
				if err == nil {
					flusher.Flush()
				}
			case exitCode := <-done:
				exitEvent := fmt.Sprintf("event: exit\ndata: %d\n\n", exitCode)

				_, err := io.WriteString(writer, exitEvent)
				if err == nil {
					flusher.Flush()
				}

				return
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

	exitCode := 0
	if results != nil {
		exitCode = results.ExitCode
	}

	done <- exitCode

	if err != nil {
		return fmt.Errorf("deploy err: %w", err)
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
