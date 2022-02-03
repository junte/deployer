package core

import (
	"bytes"
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

	go deployComponent(component, &componentConfig, args)

	return
}

func getComponent(componentName, key string) (component ComponentConfig, err error) {
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

func deployComponent(component string, config *ComponentConfig, args map[string]string) {
	command, err := prepareCommand(config.Command, args)
	if err != nil {
		log.Printf("error on prepare command: %v", err)
		return
	}

	log.Printf("exec command: %s", command)

	cmd := exec.Command(command[0], command[1:]...)
	var outBuffer, errBuffer bytes.Buffer
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

	notifyComponentDeployed(component, config, err != nil, stdout, strerr)
}

func prepareCommand(commandTemplate []string, args map[string]string) (command []string, err error) {
	context := commandContext{
		Args: args,
	}

	for _, commandItem := range commandTemplate {
		var templateBuffer bytes.Buffer
		var parsedTemplate *template.Template
		parsedTemplate, err = template.New(commandItem).Parse(commandItem)
		if err != nil {
			return
		}

		err = parsedTemplate.Execute(&templateBuffer, context)
		if err != nil {
			return
		}

		command = append(command, templateBuffer.String())
	}

	return
}
