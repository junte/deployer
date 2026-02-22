package deployer

import (
	"fmt"

	"deployer/internal/config"
	"deployer/internal/core"
)

func DeployComponent(request *core.ComponentDeployRequest) error {
	componentConfig, err := getComponentConfig(request)
	if err != nil {
		return fmt.Errorf("error on get component config: %w", err)
	}

	componentDeployer := ComponentDeployer{
		request: request,
		config:  &componentConfig,
	}

	if request.IsAsync {
		go componentDeployer.DeployAsync()
	} else {
		return componentDeployer.Deploy()
	}

	return nil
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
