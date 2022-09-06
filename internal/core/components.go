package core

import (
	"bytes"
	"deployer/internal/config"
	"deployer/internal/notify"
	"errors"
	"log"
	"os/exec"
	"text/template"
)

type commandContext struct {
	Args map[string]string
}

func DeployComponent(component, key string, args map[string]string) (err error) {
	componentConfig, err := getComponent(component, key)
	if err != nil {
		return
	}

	go internalDeployComponent(component, &componentConfig, args)

	return
}

func getComponent(componentName, key string) (component config.ComponentConfig, err error) {
	component, ok := config.Config.Components[componentName]
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

func internalDeployComponent(component string, config *config.ComponentConfig, args map[string]string) {
	command, err := prepareCommand(config.Command, args)
	if err != nil {
		log.Printf("error on prepare command: %v", err)
		return
	}

	log.Printf("exec command: %s", command)

	var outBuffer, errBuffer bytes.Buffer

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	err = cmd.Run()

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

	notify.NotifyComponentDeployed(component, config, err != nil, stdout, strerr)
}

func prepareCommand(commandTemplate []string, args map[string]string) (command []string, err error) {
	context := commandContext{
		Args: args,
	}

	for _, commandItem := range commandTemplate {
		var parsedTemplate *template.Template

		if parsedTemplate, err = template.New(commandItem).Parse(commandItem); err != nil {
			return
		}

		var templateBuffer bytes.Buffer

		if err = parsedTemplate.Execute(&templateBuffer, context); err != nil {
			return
		}

		command = append(command, templateBuffer.String())
	}

	return
}
