package server

import (
	"deployer/internal/config"
	"deployer/internal/core"
	"fmt"
	"log"
	"net/http"
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
	if config.Config.TLS.Cert != "" && config.Config.TLS.Key != "" {
		log.Printf("starting https server on port %s", config.Config.Port)
		err = http.ListenAndServeTLS(config.Config.Port, config.Config.TLS.Cert, config.Config.TLS.Key, nil)
	} else {
		log.Printf("starting http server on port %s", config.Config.Port)
		err = http.ListenAndServe(config.Config.Port, nil)
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
