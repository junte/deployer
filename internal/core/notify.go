package core

func notifyComponentDeployed(config *ComponentConfig, fail bool, output, errors string, args map[string]string) {
	go notifySlack(config, fail, output, errors, args)
}
