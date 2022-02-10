package notify

import (
	"deployer/internal/config"
)

func NotifyComponentDeployed(component string, config *config.ComponentConfig, fail bool, stdout, stderr string) {
	go notifySlack(component, config, fail, stdout, stderr)
}
