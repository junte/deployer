package core

import "deployer/src/config"

type ComponentDeployRequest struct {
	ComponentName string
	ComponentKey  string
	Args          map[string]string
	Output        *chan string
	IsAsync       bool
}

type ComponentDeployResults struct {
	StdErr   []string
	StdOut   []string
	ExitCode int
	Config   *config.ComponentConfig
	Request  *ComponentDeployRequest
}
