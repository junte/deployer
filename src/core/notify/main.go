package notify

import (
	"deployer/src/core"
	"deployer/src/core/notify/slack"
)

func NotifyComponentDeployed(results *core.ComponentDeployResults) {
	slack.Notify(results)
}
