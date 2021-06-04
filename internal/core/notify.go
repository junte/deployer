package core

func notifyComponentDeployed(config *ComponentConfig, fail bool, output, errors string) {
	go notifySlack(config, fail, output, errors)
}
