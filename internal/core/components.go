package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func DeployComponent(componentName string, key string, args map[string]string) (err error) {
	component, err := getComponent(componentName, key)
	if err != nil {
		return
	}

	command := prepareCommand(component.Command, args)
	go execCommand(command)

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

func execCommand(command []string) {
	log.Printf("exec command: %s", command)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.Printf("command err: %v", err)
	}
}
