package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"os"
	"slack-bot/app"
	"slack-bot/app/clients/aws"
	"slack-bot/app/clients/slack"
	"strconv"
)

var (
	monitoredQueues = os.Getenv("MONITORED_QUEUES")
	maxQueueDepths  = os.Getenv("MAX_QUEUE_DEPTHS")
	slackChannel    = os.Getenv("SLACK_CHANNEL")
	apiToken        = os.Getenv("SLACK_TOKEN")
	awsRegion       = os.Getenv("AWS_REGION")
	awsServer       = "sqs." + awsRegion + ".amazonaws.com"
	awsClientID     = os.Getenv("AWS_CLIENT_ID")
)

func pollQueues(slackClient *slack.SlackClient) {
	queues := []string{"PAUL_TEST", "JING_TEST"}
	awsClient := aws.CreateSqsClient(awsServer, awsRegion, awsClientID)
	m := app.Message{}
	m.Type = "message"
	m.Channel = slackChannel //"C0MFL9YTY"
	m.User = "@here"
	for {
		<-time.After(5 * time.Second)
		for _, queue := range queues {
			depth, _ := strconv.Atoi(awsClient.GetSQSQueueDepth(queue))
			timeStr := time.Now().Format("2006-01-02 15:04:05")
			m.Text = fmt.Sprintf("%s - Queue: %s, Depth %d", timeStr, queue, depth)
			err := slackClient.PostMessage(m)
			if err != nil {
				fmt.Printf("Error sending queue deoth messaage to Slack: %s\n", err.Error())
			}
		}
	}
}

func main() {
	slackClient := slack.CreateSlackClient(apiToken)
	notes := make(map[string]map[int]string)

	// start a websocket-based Real Time API session
	id := slackClient.SlackConnect()
	fmt.Println("ten-bot ready, ^C exits")

	go pollQueues(slackClient)

	for {
		// read each incoming message
		m, err := slackClient.GetMessage()
		if err != nil {
			log.Fatal(err)
		}
		// see if we're mentioned
		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			// if so try to parse if
			parts := strings.Fields(m.Text)

			if len(parts) > 2 && parts[1] == "save" {
				note := strings.Join(append(parts[:0], parts[2:]...), " ")
				aMap := map[int]string{
					0: note,
				}
				notes[m.User] = aMap
				m.Text = fmt.Sprintf("@%s - You asked me to save: %s\n", m.User, note)
			} else if parts[1] == "get" {
				aMap := notes[m.User]
				m.Text = fmt.Sprintf("@%s - Your note is: %s\n", m.User, aMap[0])
			} else {
				// huh?
				m.Text = fmt.Sprintf("@%s - sorry, that does not compute\n", m.User)
			}
			slackClient.PostMessage(m)
		}
	}
}
