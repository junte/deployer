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
	component, ok := Config.Components[componentName]
	if !ok {
		err = errors.New("component not found")
		return
	}

	if component.Key != key {
		err = errors.New("keys mismatch")
		return
	}

	var commandArgs []string
	for argKey, argValue := range args {
		commandArgs = append(commandArgs, fmt.Sprintf("${arg_%s}", argKey))
		commandArgs = append(commandArgs, argValue)
	}

	replacer := strings.NewReplacer(commandArgs...)
	command := replacer.Replace(component.Command)

	go runCommand(command)

	return
}

func runCommand(command string) {
	log.Printf("command: %s", command)

	cmd := &exec.Cmd{
		Path:   "/bin/bash",
		Args:   []string{"/bin/bash", "-c", command},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	var err = cmd.Run()
	if err != nil {
		log.Printf("command err: %v", err)
	}
}
