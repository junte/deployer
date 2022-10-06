package notify

import (
	"deployer/internal/core"
	"deployer/internal/core/notify/slack"
)

func NotifyComponentDeployed(results *core.ComponentDeployResults) {
	slack.Notify(results)
}
