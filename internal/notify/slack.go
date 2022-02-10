package notify

import (
	"deployer/internal/config"
	"fmt"
	"github.com/slack-go/slack"
	"log"
)

func notifySlack(component string, componentConfig *config.ComponentConfig, failed bool, stdout, stderr string) {
	if config.Config.Notification.Slack.ApiToken == "" {
		return
	}

	channel := getSlackChannel(componentConfig)
	if channel == "" {
		return
	}

	slackMessage := buildSlackMessage(component, failed, stdout, stderr)

	client := slack.New(config.Config.Notification.Slack.ApiToken)
	_, _, err := client.PostMessage(channel, slackMessage)
	if err != nil {
		log.Printf("error on slack message send: %v", err)
		return
	}
}

func buildSlackMessage(component string, failed bool, stdout, stderr string) slack.MsgOption {
	message := ""
	if failed {
		message = fmt.Sprintf(":x: Failed component \"%s\" deployment to environment \"%s\"",
			component,
			config.Config.Environment,
		)
	} else {
		message = fmt.Sprintf(":white_check_mark: Component \"%s\" was deployed to environment \"%s\"",
			component,
			config.Config.Environment,
		)
	}

	attachments := []slack.Attachment{}
	attachments = append(attachments, slack.Attachment{
		Title:   ":memo: stdout",
		Pretext: message,
		Color:   "#36a64f",
		Text:    stdout,
	})
	if stderr != "" {
		attachments = append(attachments, slack.Attachment{
			Title: ":fire: strerr",
			Color: "#eb343a",
			Text:  stderr,
		})
	}
	return slack.MsgOptionAttachments(attachments...)
}

func getSlackChannel(componentConfig *config.ComponentConfig) string {
	channel := config.Config.Notification.Slack.Channel
	if componentConfig.Notification.Slack.Channel != "" {
		channel = componentConfig.Notification.Slack.Channel
	}

	return channel
}
