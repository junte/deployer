package deployer

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"deployer/src/config"
	"deployer/src/core"
)

func TestPrepareCommand(t *testing.T) {
	var tests = []struct {
		command []string
		args    map[string]string
		want    []string
	}{
		{
			[]string{"/bin/bash", "-c", "./internalDeploy --tag={{.Args.tag}}"},
			map[string]string{
				"tag":     "124",
				"command": "rm -rf /",
			},
			[]string{"/bin/bash", "-c", "./internalDeploy --tag=124"}},
		{
			[]string{"/bin/bash", "-c", "echo Hello World"},
			map[string]string{"command": "rm -rf /"},
			[]string{"/bin/bash", "-c", "echo Hello World"},
		},
		{
			[]string{"/run.sh"},
			map[string]string{},
			[]string{"/run.sh"},
		},
	}

	dep := ComponentDeployer{}

	for _, testCase := range tests {
		t.Run(fmt.Sprintf("%s,%s", testCase.command, testCase.args), func(t *testing.T) {
			command, err := dep.prepareCommand(testCase.command, testCase.args)
			if err != nil {
				t.Errorf("err: %s", err)
			} else if !reflect.DeepEqual(command, testCase.want) {
				t.Errorf("got %s, want %s", command, testCase.want)
			}
		})
	}
}

func TestInternalDeployCapturesOutput(t *testing.T) {
	tests := []struct {
		name         string
		command      []string
		wantExitCode int
		wantStdOut   []string
		wantStdErr   []string
	}{
		{
			name:         "captures stdout with zero exit",
			command:      []string{"/bin/sh", "-c", "echo hello"},
			wantExitCode: 0,
			wantStdOut:   []string{"hello\n"},
			wantStdErr:   nil,
		},
		{
			name:         "captures stderr with non-zero exit",
			command:      []string{"/bin/sh", "-c", "echo oops 1>&2; exit 2"},
			wantExitCode: 2,
			wantStdOut:   nil,
			wantStdErr:   []string{"oops\n"},
		},
		{
			name:         "captures both streams",
			command:      []string{"/bin/sh", "-c", "echo out; echo err 1>&2; exit 3"},
			wantExitCode: 3,
			wantStdOut:   []string{"out\n"},
			wantStdErr:   []string{"err\n"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			componentDeployer := ComponentDeployer{
				request: &core.ComponentDeployRequest{},
				config: &config.ComponentConfig{
					Command: testCase.command,
					Timeout: 30 * time.Second,
				},
			}

			results, err := componentDeployer.internalDeploy()
			if err != nil {
				t.Fatalf("err: %s", err)
			}

			if results.ExitCode != testCase.wantExitCode {
				t.Errorf("exit code: got %d, want %d", results.ExitCode, testCase.wantExitCode)
			}

			if !reflect.DeepEqual(results.StdOut, testCase.wantStdOut) {
				t.Errorf("stdout: got %q, want %q", results.StdOut, testCase.wantStdOut)
			}

			if !reflect.DeepEqual(results.StdErr, testCase.wantStdErr) {
				t.Errorf("stderr: got %q, want %q", results.StdErr, testCase.wantStdErr)
			}
		})
	}
}

func TestResolveTimeout(t *testing.T) {
	originalConfig := config.Config

	defer func() {
		config.Config = originalConfig
	}()

	tests := []struct {
		name             string
		globalTimeout    time.Duration
		componentTimeout time.Duration
		want             time.Duration
	}{
		{
			name:             "component timeout overrides global",
			globalTimeout:    10 * time.Minute,
			componentTimeout: 30 * time.Minute,
			want:             30 * time.Minute,
		},
		{
			name:             "falls back to global when component timeout is zero",
			globalTimeout:    10 * time.Minute,
			componentTimeout: 0,
			want:             10 * time.Minute,
		},
		{
			name:             "uses global timeout when no component override",
			globalTimeout:    5 * time.Minute,
			componentTimeout: 0,
			want:             5 * time.Minute,
		},
		{
			name:             "short component timeout",
			globalTimeout:    10 * time.Minute,
			componentTimeout: 30 * time.Second,
			want:             30 * time.Second,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config.Config.Timeout = testCase.globalTimeout

			componentConfig := &config.ComponentConfig{
				Timeout: testCase.componentTimeout,
			}

			dep := ComponentDeployer{
				config: componentConfig,
			}

			got := dep.resolveTimeout()
			if got != testCase.want {
				t.Errorf("got %s, want %s", got, testCase.want)
			}
		})
	}
}
