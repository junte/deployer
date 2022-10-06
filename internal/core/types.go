package core

import "deployer/internal/config"

type ComponentDeployRequest struct {
	ComponentName string
	ComponentKey  string
	Args          map[string]string
	Output        *chan string
	IsAsync       bool
}

type ComponentDeployResults struct {
	StdErr  []string
	StdOut  []string
	Config  *config.ComponentConfig
	Request *ComponentDeployRequest
}
