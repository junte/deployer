package deployer

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"deployer/src/config"
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
