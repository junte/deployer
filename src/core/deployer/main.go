package deployer

import (
	"fmt"

	"deployer/src/config"
	"deployer/src/core"
)

func NewComponentDeployer(request *core.ComponentDeployRequest) (*ComponentDeployer, error) {
	componentConfig, err := getComponentConfig(request)
	if err != nil {
		return nil, fmt.Errorf("get component config: %w", err)
	}

	return &ComponentDeployer{
		request: request,
		config:  &componentConfig,
	}, nil
}

func getComponentConfig(request *core.ComponentDeployRequest) (config.ComponentConfig, error) {
	component, ok := config.Config.Components[request.ComponentName]
	if !ok {
		return config.ComponentConfig{}, fmt.Errorf("component not found: %s", request.ComponentName)
	}

	if component.Key != "" && component.Key != request.ComponentKey {
		return config.ComponentConfig{}, fmt.Errorf("invalid component key for component: %s", request.ComponentName)
	}

	return component, nil
}
