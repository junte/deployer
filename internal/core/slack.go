package core

import (
	"fmt"
	"github.com/slack-go/slack"
	"log"
)

func notifySlack(component_config *ComponentConfig, failed bool, output, errors string, args map[string]string) {
	if Config.Notification.Slack.ApiToken == "" {
		return
	}

	channel := getSlackChannel(component_config)
	if channel == "" {
		return
	}

	slackMessage := buildSlackMessage(
		component_config,
		failed,
		output,
		errors,
	)

	client := slack.New(Config.Notification.Slack.ApiToken, slack.OptionDebug(true))
	_, _, err := client.PostMessage(channel, slackMessage)
	if err != nil {
		log.Printf("error on slack message send: %v", err)
		return
	}
}

func buildSlackMessage(componentConfig *ComponentConfig, failed bool, output, errors string) slack.MsgOption {
	message := ""
	if failed {
		message = fmt.Sprintf(":x: Failed component \"%s\" deployment to environment \"%s\"",
			componentConfig.Key,
			Config.Environment,
		)
	} else {
		message = fmt.Sprintf(":white_check_mark: Component \"%s\" was deployed to environment \"%s\"",
			componentConfig.Key,
			Config.Environment,
		)
	}

	attachments := []slack.Attachment{}
	attachments = append(attachments, slack.Attachment{
		Title:   ":memo: Log",
		Pretext: message,
		Color:   "#36a64f",
		Text:    output,
	})
	if errors != "" {
		attachments = append(attachments, slack.Attachment{
			Title: ":fire: Errors",
			Color: "#eb343a",
			Text:  errors,
		})
	}
	return slack.MsgOptionAttachments(attachments...)
}

func getSlackChannel(component_config *ComponentConfig) string {
	channel := Config.Notification.Slack.Channel
	if component_config.Notification.Slack.Channel != "" {
		channel = component_config.Notification.Slack.Channel
	}

	return channel
}
