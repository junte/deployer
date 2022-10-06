package deployer

import (
	"deployer/internal/config"
	"deployer/internal/core"
	"errors"
)

func DeployComponent(request *core.ComponentDeployRequest) (err error) {
	componentConfig, err := getComponentConfig(request)
	if err != nil {
		return
	}

	componentDeployer := ComponentDeployer{
		request: request,
		config:  &componentConfig,
	}

	if request.IsAsync {
		go componentDeployer.Deploy()
	} else {
		return componentDeployer.Deploy()
	}

	return
}

func getComponentConfig(request *core.ComponentDeployRequest) (component config.ComponentConfig, err error) {
	component, ok := config.Config.Components[request.ComponentName]
	if !ok {
		err = errors.New("component not found")
		return
	}

	if component.Key != "" && component.Key != request.ComponentKey {
		err = errors.New("keys mismatch")
		return
	}

	return
}
