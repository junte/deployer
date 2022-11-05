package slack

import (
	"deployer/internal/config"
	"deployer/internal/core"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"strings"
)

func Notify(results *core.ComponentDeployResults) {
	if config.Config.Notification.Slack.ApiToken == "" {
		return
	}

	channel := getSlackChannel(results.Config)
	if channel == "" {
		return
	}

	slackMessage := buildSlackMessage(results)

	client := slack.New(config.Config.Notification.Slack.ApiToken)
	if _, _, err := client.PostMessage(channel, slackMessage); err != nil {
		log.WithError(err).Error("error on slack message send")
		return
	}
}

func buildSlackMessage(results *core.ComponentDeployResults) slack.MsgOption {
	message := ""
	if results.ExitCode == 0 {
		message = fmt.Sprintf(":white_check_mark: Component \"%s\" was deployed to environment \"%s\"",
			results.Request.ComponentName,
			config.Config.Environment,
		)
	} else {
		message = fmt.Sprintf(":x: Failed component \"%s\" deployment to environment \"%s\"",
			results.Request.ComponentName,
			config.Config.Environment,
		)
	}

	var attachments []slack.Attachment
	attachments = append(attachments, slack.Attachment{
		Title:   ":memo: stdout",
		Pretext: message,
		Color:   "#36a64f",
		Text:    strings.Join(results.StdOut, "\n"),
	})

	if len(results.StdErr) > 0 {
		attachments = append(attachments, slack.Attachment{
			Title: ":fire: stderr",
			Color: "#eb343a",
			Text:  strings.Join(results.StdErr, "\n"),
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
