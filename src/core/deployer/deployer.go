package deployer

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"text/template"
	"time"

	"deployer/src/config"
	"deployer/src/core"
	"deployer/src/core/notify"

	log "github.com/sirupsen/logrus"
)

type commandTemplateContext struct {
	Args map[string]string
}

type ComponentDeployer struct {
	request *core.ComponentDeployRequest
	config  *config.ComponentConfig
}

func (deployer *ComponentDeployer) Deploy() (*core.ComponentDeployResults, error) {
	results, err := deployer.internalDeploy()
	if err != nil {
		return nil, fmt.Errorf("error on deploy component: %w", err)
	}

	go notify.NotifyComponentDeployed(results)

	return results, nil
}

func (deployer *ComponentDeployer) DeployAsync() {
	results, err := deployer.internalDeploy()
	if err != nil {
		return
	}

	go notify.NotifyComponentDeployed(results)
}

func (deployer *ComponentDeployer) internalDeploy() (*core.ComponentDeployResults, error) {
	command, err := deployer.prepareCommand(deployer.config.Command, deployer.request.Args)
	if err != nil {
		return nil, fmt.Errorf("error on prepare command: %w", err)
	}

	timeout := deployer.resolveTimeout()
	ctx, contextCancel := context.WithTimeout(context.Background(), timeout)

	defer contextCancel()

	log.Debugf("exec command: %s (timeout: %s)", command, timeout)
	cmd := exec.CommandContext(ctx, command[0], command[1:]...) //nolint:gosec
	cmd.Dir = deployer.config.WorkDir

	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		log.WithError(err).Error("failed creating command cmdStdout pipe")
		return nil, fmt.Errorf("error creating stdout pipe: %w", err)
	}

	defer func() {
		_ = cmdStdout.Close()
	}()

	cmdStderr, err := cmd.StderrPipe()
	if err != nil {
		log.WithError(err).Error("failed creating command cmdStderr pipe")
		return nil, fmt.Errorf("error creating stderr pipe: %w", err)
	}

	defer func() {
		_ = cmdStderr.Close()
	}()

	stdoutReader := bufio.NewReader(cmdStdout)
	stderrReader := bufio.NewReader(cmdStderr)

	err = cmd.Start()
	if err != nil {
		log.WithError(err).Error("error starting command")
		return nil, fmt.Errorf("error starting command: %w", err)
	}

	stdout := make(chan string)
	stderr := make(chan string)

	var (
		stdoutLines []string
		stderrLines []string
	)

	var aggregateWaitGroup sync.WaitGroup

	aggregateWaitGroup.Go(func() {
		deployer.aggregateOutput(stdout, stderr, &stdoutLines, &stderrLines)
	})

	var readersWaitGroup sync.WaitGroup

	readersWaitGroup.Go(func() {
		deployer.handleReader(stdout, stdoutReader)
	})
	readersWaitGroup.Go(func() {
		deployer.handleReader(stderr, stderrReader)
	})

	// wait for both pipes to reach EOF (process exit or ctx timeout kills the
	// process and closes the pipes) before closing the internal channels, so no
	// reader ever sends on a closed channel.
	readersWaitGroup.Wait()
	close(stdout)
	close(stderr)

	// wait for aggregation to drain both channels before reading the accumulated
	// lines below and before returning, so forwarding to the request channels and
	// the slice appends are complete.
	aggregateWaitGroup.Wait()

	var exitCode int

	err = cmd.Wait()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Warnf("command timed out after %s", timeout)

			return &core.ComponentDeployResults{
				Request:  deployer.request,
				Config:   deployer.config,
				StdErr:   append(stderrLines, fmt.Sprintf("deployment timed out after %s\n", timeout)),
				StdOut:   stdoutLines,
				ExitCode: 1,
			}, nil
		}

		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			if status, isWaitStatus := exitErr.Sys().(syscall.WaitStatus); isWaitStatus {
				exitCode = status.ExitStatus()
			}
		}
	}

	log.Debugf("exit status: %v", exitCode)

	deployResults := &core.ComponentDeployResults{
		Request:  deployer.request,
		Config:   deployer.config,
		StdErr:   stderrLines,
		StdOut:   stdoutLines,
		ExitCode: exitCode,
	}

	return deployResults, nil
}

func (deployer *ComponentDeployer) resolveTimeout() time.Duration {
	if deployer.config.Timeout > 0 {
		return deployer.config.Timeout
	}

	return config.Config.Timeout
}

func (*ComponentDeployer) handleReader(
	output chan<- string,
	reader *bufio.Reader,
) {
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			output <- line
		}

		if err != nil {
			return
		}
	}
}

func (deployer *ComponentDeployer) aggregateOutput(
	stdout <-chan string,
	stderr <-chan string,
	stdoutLines *[]string,
	stderrLines *[]string,
) {
	for stdout != nil || stderr != nil {
		select {
		case line, ok := <-stdout:
			if !ok {
				stdout = nil

				continue
			}

			*stdoutLines = append(*stdoutLines, line)

			if deployer.request.Output != nil {
				*deployer.request.Output <- line
			}

			log.Debug(line)
		case line, ok := <-stderr:
			if !ok {
				stderr = nil

				continue
			}

			*stderrLines = append(*stderrLines, line)

			if deployer.request.ErrorOutput != nil {
				*deployer.request.ErrorOutput <- line
			}

			log.Debug(line)
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
