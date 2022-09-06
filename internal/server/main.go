package server

import (
	"deployer/internal/config"
	"deployer/internal/core"
	"fmt"
	"log"
	"net/http"
	"time"
)

func Run() {
	log.Printf("version: %s", config.Version)

	config.ReadConfig()

	http.HandleFunc("/", handler)

	err := startServer()
	if err != nil {
		log.Printf("failed start server: %s", err)
	}
}

func startServer() (err error) {
	server := &http.Server{
		Addr:              config.Config.Port,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if config.Config.TLS.Cert != "" && config.Config.TLS.Key != "" {
		log.Printf("starting https server on port %s", config.Config.Port)
		err = server.ListenAndServeTLS(config.Config.TLS.Cert, config.Config.TLS.Key)
	} else {
		log.Printf("starting http server on port %s", config.Config.Port)
		err = server.ListenAndServe()
	}

	return
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

	args := make(map[string]string)
	for key, values := range request.Form {
		args[key] = values[0]
	}

	err := core.DeployComponent(request.FormValue("component"), request.FormValue("key"), args)
	if err != nil {
		http.Error(writer, fmt.Sprintf("deploy err: %v", err), http.StatusBadRequest)
		return
	}
}
