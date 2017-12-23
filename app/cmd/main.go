package main

import (
	"fmt"
	"strings"
	"time"

	"os"
	"slack-bot/app"
	"slack-bot/app/clients/aws"
	"github.com/nlopes/slack"
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

func outputMessage(queueList []app.Queue) {
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	params := slack.PostMessageParameters{}

	numAlertQueues := 0
	text := "```"
	text = text + fmt.Sprintf("%-50s%-20s%s\n","Queue", "Current Depth", "Alert Depth")
	text = text + fmt.Sprintln("___")
	for x:=0;x<len(queueList);x++ {
		if queueList[x].LastDepth >= queueList[x].MaxDepth {
			numAlertQueues++
			text = text + fmt.Sprintf("%-50s%-20d%d\n", queueList[x].Name, queueList[x].LastDepth, queueList[x].MaxDepth)
		}
	}
	text = text + "```"
	color := "#CC0000"
	if numAlertQueues == 0 {
		text = "No Queues are in trouble at this time..."
		color = "#00CC00"
	}
	attachment := slack.Attachment{
		//AuthorIcon: "https://clouda-assets.s3.amazonaws.com/upload/54d0e623d287c266052be732.png?1422976549",
		//AuthorName: "AWS Queue Depth Alert Bot",
		Color:      color,
		Title:      "Queues that exceed alert depth",
		MarkdownIn: []string{"text", "pretext"},
		Text: 	text,
		//Fields: fields,
	}
	params.Username = "AWS Queue Depth Alert Bot"
	params.AsUser = true
	params.IconEmoji = "https://avatars.slack-edge.com/2017-12-22/290637412467_5a6d58b2f73076443114_48.png"
	params.Attachments = []slack.Attachment{attachment}
	channelID, timestamp, err := api.PostMessage(os.Getenv("SLACK_CHANNEL"), "", params)
	if err != nil {
		fmt.Printf("-->>%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
}

func pollQueues(awsClient *aws.SqsClient, queueList []app.Queue) {
	var err error
	m := app.Message{}
	m.Type = "message"
	m.Channel = slackChannel
	m.User = "@here"
	for {
		for x:=0;x<len(queueList);x++ {
			queueList[x].LastDepth, err = strconv.Atoi(awsClient.GetSQSQueueDepth(queueList[x].Name))
			if err != nil {
				fmt.Println("Error:", err.Error())
			}
		}
		fmt.Printf("%+v", queueList)
		outputMessage(queueList)
		<-time.After(10 * time.Minute)
	}
}

func main() {
	queueList := make([]app.Queue, 0)
	queues := os.Getenv("QUEUES_MONITORED")
	fmt.Println(queues)
	result := strings.Split(queues, ",")
	for _, res := range result {
		tmpQ := strings.Split(res, ":")
		max, _ := strconv.Atoi(tmpQ[1])
		queue := app.Queue{
			Name:     tmpQ[0],
			MaxDepth: max,
		}
		queueList = append(queueList, queue)
	}
	fmt.Printf("%+v", queueList)

	fmt.Println("ten-bot ready, ^C exits")

	awsClient := aws.CreateSqsClient(awsServer, awsRegion, awsClientID)

	go pollQueues(awsClient, queueList)

	for {
		<-time.After(10 * time.Minute)
	}
}
