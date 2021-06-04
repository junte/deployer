package core

import (
	"fmt"
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
	ApiToken string
	Channel  string
}

type ComponentConfig struct {
	Command      []string
	Key          string
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

// Version of application
var Version = "development"

// Config of application runtime
var Config AppConfig

func ReadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		panic(fmt.Errorf("fatal error bad config file: %s", err))
	}
}
