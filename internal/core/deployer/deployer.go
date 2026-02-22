package deployer

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"text/template"

	"deployer/internal/config"
	"deployer/internal/core"
	"deployer/internal/core/notify"

	log "github.com/sirupsen/logrus"
)

type commandTemplateContext struct {
	Args map[string]string
}

type ComponentDeployer struct {
	request *core.ComponentDeployRequest
	config  *config.ComponentConfig
}

func (deployer *ComponentDeployer) Deploy() error {
	results, err := deployer.internalDeploy()
	if err != nil {
		return fmt.Errorf("error on deploy component: %w", err)
	}

	go notify.NotifyComponentDeployed(results)

	return nil
}

func (deployer *ComponentDeployer) DeployAsync() {
	results, err := deployer.internalDeploy()
	if err != nil {
		return
	}

	go notify.NotifyComponentDeployed(results)
}

func (deployer *ComponentDeployer) internalDeploy() (deployResults *core.ComponentDeployResults, err error) {
	command, err := deployer.prepareCommand(deployer.config.Command, deployer.request.Args)
	if err != nil {
		err = fmt.Errorf("error on prepare command: %w", err)
		return
	}

	log.Debugf("exec command: %s", command)
	cmd := exec.Command(command[0], command[1:]...) //nolint:gosec
	cmd.Dir = deployer.config.WorkDir

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

	defer close(stderr)
	defer close(stdout)

	context, contextCancel := context.WithCancel(context.Background())

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
			case <-context.Done():
				return
			}
		}
	}()

	go deployer.handleReader(context, &stdout, stdoutReader)
	go deployer.handleReader(context, &stderr, stderrReader)

	var exitCode int

	if err = cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if status, isWaitStatus := exitErr.Sys().(syscall.WaitStatus); isWaitStatus {
				exitCode = status.ExitStatus()
			}
		}
	}

	contextCancel()

	log.Debugf("exit status: %v", exitCode)

	deployResults = &core.ComponentDeployResults{
		Request:  deployer.request,
		Config:   deployer.config,
		StdErr:   stderrLines,
		StdOut:   stdoutLines,
		ExitCode: exitCode,
	}

	return
}

func (*ComponentDeployer) handleReader(
	context context.Context,
	output *chan string,
	reader *bufio.Reader,
) {
	for {
		select {
		case <-context.Done():
			return
		default:
			str, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			*output <- str
		}
	}
}

func (*ComponentDeployer) prepareCommand(
	commandTemplate []string,
	args map[string]string,
) ([]string, error) {
	context := commandTemplateContext{
		Args: args,
	}

	var command []string

	for _, commandItem := range commandTemplate {
		parsedTemplate, err := template.New(commandItem).Parse(commandItem)
		if err != nil {
			return nil, fmt.Errorf("error on parse command template: %w", err)
		}

		var templateBuffer bytes.Buffer

		if err = parsedTemplate.Execute(&templateBuffer, context); err != nil {
			return nil, fmt.Errorf("error on execute command template: %w", err)
		}

		command = append(command, templateBuffer.String())
	}

	return command, nil
}
