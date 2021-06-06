package core

func notifyComponentDeployed(component string, config *ComponentConfig, fail bool, stdout, stderr string) {
	go notifySlack(component, config, fail, stdout, stderr)
}
