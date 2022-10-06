package deployer

import (
	"bufio"
	"bytes"
	"deployer/internal/config"
	"deployer/internal/core"
	"deployer/internal/core/notify"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"syscall"
	"text/template"
)

type commandTemplateContext struct {
	Args map[string]string
}

type ComponentDeployer struct {
	request *core.ComponentDeployRequest
	config  *config.ComponentConfig
}

func (deployer *ComponentDeployer) Deploy() (err error) {
	results, err := deployer.deploy()
	if err != nil {
		return
	}

	notify.NotifyComponentDeployed(results)

	return
}

func (deployer *ComponentDeployer) DeployAsync() {
	results, err := deployer.deploy()
	if err != nil {
		return
	}

	notify.NotifyComponentDeployed(results)
}

func (deployer *ComponentDeployer) deploy() (deployResults *core.ComponentDeployResults, err error) {
	command, err := deployer.prepareCommand(deployer.config.Command, deployer.request.Args)

	if err != nil {
		err = fmt.Errorf("error on prepare command: %v", err)
		return
	}

	log.Debugf("exec command: %s", command)
	cmd := exec.Command(command[0], command[1:]...)
	cmdStdout, err := cmd.StdoutPipe()

	if err != nil {
		log.WithError(err).Error("failed creating command cmdStdout pipe")
		return
	}

	defer func() {
		_ = cmdStdout.Close()
	}()

	cmdStderr, err := cmd.StderrPipe()
	if err != nil {
		log.WithError(err).Error("failed creating command cmdStderr pipe")
		return
	}

	defer func() {
		_ = cmdStderr.Close()
	}()

	stdoutReader := bufio.NewReader(cmdStdout)
	stderrReader := bufio.NewReader(cmdStderr)

	if err = cmd.Start(); err != nil {
		log.WithError(err).Error("failed starting command")
		return
	}

	stdout := make(chan string)
	stderr := make(chan string)
	done := make(chan bool)

	defer close(stderr)
	defer close(stdout)
	defer close(done)

	var (
		stdoutLines []string
		stderrLines []string
	)

	go func() {
		for {
			select {
			case line, more := <-stdout:
				if !more {
					break
				}

				stdoutLines = append(stdoutLines, line)

				if deployer.request.Output != nil {
					*deployer.request.Output <- line
				}

				log.Debug(line)
			case line, more := <-stderr:
				if !more {
					break
				}

				stderrLines = append(stderrLines, line)

				if deployer.request.Output != nil {
					*deployer.request.Output <- line
				}

				log.Debug(line)
			case <-done:
				break
			}
		}
	}()

	go deployer.handleReader(&stdout, stdoutReader)
	go deployer.handleReader(&stderr, stderrReader)

	if err = cmd.Wait(); err != nil {
		if exiterr, isExitErr := err.(*exec.ExitError); isExitErr {
			if status, isWaitStatus := exiterr.Sys().(syscall.WaitStatus); isWaitStatus {
				log.Debugf("exit status: %v", status.ExitStatus())
			}
		}
	}

	done <- true

	deployResults = &core.ComponentDeployResults{
		Request: deployer.request,
		Config:  deployer.config,
		StdErr:  stderrLines,
		StdOut:  stdoutLines,
	}

	return
}

func (deployer *ComponentDeployer) handleReader(output *chan string, reader *bufio.Reader) {
	for {
		str, err := reader.ReadString('\n')
		if len(str) == 0 && err != nil {
			if err == io.EOF {
				break
			}
		}

		*output <- str
	}
}

func (deployer *ComponentDeployer) prepareCommand(commandTemplate []string, args map[string]string) (command []string, err error) {
	context := commandTemplateContext{
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