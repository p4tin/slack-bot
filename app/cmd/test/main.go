package main

import (
	"fmt"

	"os"

	"github.com/nlopes/slack"
)

func main() {
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		AuthorName: "queuedepthbot",
		Color:      "#CC0000",
		Title:      "Queues that exceeed max limits",
		Text:       "*For more info see the AWS Console:* https://console.aws.amazon.com/console/home\nPosted at: {date_short} @ {time_secs}",
		// Uncomment the following part to send a field too
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: "",
				Short: true,
			},
			slack.AttachmentField{
				Title: "",
				Short: true,
			},
			slack.AttachmentField{
				Title: "Queue Name",
				Short: true,
			},
			slack.AttachmentField{
				Title: "Depth",
				Short: true,
			},
			slack.AttachmentField{
				Value: "A15-PROD1-INVENTORY-QUEUE",
				Short: true,
			},
			slack.AttachmentField{
				Title: "125",
				Short: true,
			},
			slack.AttachmentField{
				Value: "A15-PROD1-PRODUCT-UPDATE-QUEUE",
				Short: true,
			},
			slack.AttachmentField{
				Title: "145",
				Short: true,
			},
		},
	}
	params.Attachments = []slack.Attachment{attachment}
	channelID, timestamp, err := api.PostMessage(os.Getenv("SLACK_CHANNEL"), "", params)
	if err != nil {
		fmt.Printf("-->>%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
