package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

type config struct {
	Port       string
	Components map[string]componentConfig
}

type componentConfig struct {
	Command string
	Key     string
}

// Version of application
var Version = "development"
var appConfig config

func main() {
	log.Printf("Version %s", Version)

	readConfig()

	http.HandleFunc("/", handler)

	log.Printf("Starting server on port %s", appConfig.Port)

	err := http.ListenAndServeTLS(appConfig.Port, "./tls/cert.crt", "./tls/cert.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // optionally look for config in the working directory

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	err = viper.Unmarshal(&appConfig)
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error bad config file: %s", err))
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("ParseForm() err: %v", err), http.StatusBadRequest)
		return
	}

	args := make(map[string]string)
	for key, values := range r.Form {
		args[key] = values[0]
	}

	err := deployComponent(r.FormValue("component"), r.FormValue("key"), args)
	if err != nil {
		http.Error(w, fmt.Sprintf("deploy err: %v", err), http.StatusBadRequest)
		return
	}
}

func deployComponent(componentName string, key string, args map[string]string) (err error) {
	component, ok := appConfig.Components[componentName]
	if !ok {
		err = errors.New("component not found")
		return
	}

	if component.Key != key {
		err = errors.New("keys mismatch")
		return
	}

	commandArgs := []string{}
	for argKey, argValue := range args {
		commandArgs = append(commandArgs, fmt.Sprintf("${arg_%s}", argKey))
		commandArgs = append(commandArgs, argValue)
	}

	replacer := strings.NewReplacer(commandArgs...)
	command := replacer.Replace(component.Command)

	go runCommand(command)

	return
}

func runCommand(command string) (err error) {
	log.Printf("command: %s", command)

	cmd := &exec.Cmd{
		Path:   "/bin/bash",
		Args:   []string{"/bin/bash", "-c", command},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	err = cmd.Run()
	if err != nil {
		log.Printf("Fatal error config file: %s", err)
	}

	return
}
