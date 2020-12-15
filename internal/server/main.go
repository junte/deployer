package server

import (
	"deployer/internal/core"
	"fmt"
	"log"
	"net/http"
)

func Run() {
	log.Printf("version: %s", core.Version)

	core.ReadConfig()

	http.HandleFunc("/", handler)

	err := startServer()
	if err != nil {
		log.Fatal("failed start server: ", err)
	}
}

func startServer() (err error) {
	if core.Config.TLS.Cert != "" && core.Config.TLS.Key != "" {
		log.Printf("starting https server on port %s", core.Config.Port)
		err = http.ListenAndServeTLS(core.Config.Port, core.Config.TLS.Cert, core.Config.TLS.Key, nil)
	} else {
		log.Printf("starting http server on port %s", core.Config.Port)
		err = http.ListenAndServe(core.Config.Port, nil)
	}

	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("wrong query params err: %v", err), http.StatusBadRequest)
		return
	}

	args := make(map[string]string)
	for key, values := range r.Form {
		args[key] = values[0]
	}

	err := core.DeployComponent(r.FormValue("component"), r.FormValue("key"), args)
	if err != nil {
		http.Error(w, fmt.Sprintf("deploy err: %v", err), http.StatusBadRequest)
		return
	}
}
