package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Port         string
	Environment  string
	Components   map[string]ComponentConfig
	TLS          TLSConfig
	Notification NotificationConfig
}

type NotificationConfig struct {
	Slack SlackConfig
}

type SlackConfig struct {
	// nolint:gosec
	ApiToken string
	Channel  string
}

type ComponentConfig struct {
	Command      []string
	Key          string
	WorkDir      string
	Notification ComponentNotificationConfig
}

type ComponentNotificationConfig struct {
	Slack SlackComponentConfig
}

type SlackComponentConfig struct {
	Channel string
}

type TLSConfig struct {
	Cert string
	Key  string
}

// Version of application.
var Version = "development"

// Config of application runtime.
var Config AppConfig

func ReadConfig(configFile string) error {
	configName := strings.Split(configFile, ".")[0]

	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	return nil
}
