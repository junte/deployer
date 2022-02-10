package notify

import "deployer/internal/core"

func NotifyComponentDeployed(component string, config *core.ComponentConfig, fail bool, stdout, stderr string) {
	go notifySlack(component, config, fail, stdout, stderr)
}
