package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func DeployComponent(component string, key string, args map[string]string) (err error) {
	componentConfig, err := getComponent(component, key)
	if err != nil {
		return
	}

	go deployComponent(component, &componentConfig, args)

	return
}

func getComponent(componentName string, key string) (component ComponentConfig, err error) {
	component, ok := Config.Components[componentName]
	if !ok {
		err = errors.New("component not found")
		return
	}

	if component.Key != "" && component.Key != key {
		err = errors.New("keys mismatch")
		return
	}
	return
}

func prepareCommand(command []string, args map[string]string) []string {
	var commandArgs []string
	for argKey, argValue := range args {
		commandArgs = append(commandArgs, fmt.Sprintf("${arg_%s}", argKey))
		commandArgs = append(commandArgs, argValue)
	}

	replacer := strings.NewReplacer(commandArgs...)

	var preparedCommand []string
	for _, commandItem := range command {
		preparedCommand = append(preparedCommand, replacer.Replace(commandItem))
	}

	return preparedCommand
}

func deployComponent(component string, componentConfig *ComponentConfig, args map[string]string) {
	command := prepareCommand(componentConfig.Command, args)

	log.Printf("exec command: %s", command)

	cmd := exec.Command(command[0], command[1:]...)
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	err := cmd.Run()

	if err != nil {
		log.Printf("error on run command: %v", err)
	}

	stdout := outBuffer.String()
	strerr := errBuffer.String()

	if stdout != "" {
		log.Printf("command stdout:\n%v", stdout)
	}

	if strerr != "" {
		log.Printf("command error:\n%v", strerr)
	}

	notifyComponentDeployed(component, componentConfig, err != nil, stdout, strerr)
}
