package main

import (
	"fmt"
	"log"
	"os"
	"slack-bot/app"
	"slack-bot/app/clients/aws"
	lslack "slack-bot/app/clients/slack"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
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
	t := time.Now()
	timeStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d-00:00\n",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	slackClient := slack.New(os.Getenv("SLACK_TOKEN"))
	params := slack.PostMessageParameters{}

	numAlertQueues := 0
	text := "```"
	text = text + fmt.Sprintf("%-50s%-15s%s\n", "Queue", "Current Depth", "Alert Depth")
	text = text + fmt.Sprintln("___")
	for x := 0; x < len(queueList); x++ {
		if (queueList[x].LastDepth >= queueList[x].MaxDepth) && !queueList[x].Paused {
			numAlertQueues++
		}
		if !queueList[x].Paused {
			text = text + fmt.Sprintf("%-50s%-15d%d\n", queueList[x].Name, queueList[x].LastDepth, queueList[x].MaxDepth)
		}
	}
	text = text + "```"
	color := "#CC0000"
	title := fmt.Sprintf("%d queue(s) exceed the alert depth", numAlertQueues)
	if numAlertQueues == 0 {
		title = "No queue(s) exceed the alert depth at this time"
		color = "#00CC00"
	}

	attachment := slack.Attachment{
		Pretext:    timeStr + "\n*For help type `@QueueDepthBot help`*",
		Color:      color,
		Title:      title,
		MarkdownIn: []string{"text", "pretext"},
		Text:       text,
	}
	params.Username = "AWS Queue Depth Alert Bot"
	params.AsUser = true
	params.IconEmoji = "https://avatars.slack-edge.com/2017-12-22/290637412467_5a6d58b2f73076443114_48.png"
	params.Attachments = []slack.Attachment{attachment}
	channelID, timestamp, err := slackClient.PostMessage(os.Getenv("SLACK_CHANNEL"), "", params)
	if err != nil {
		fmt.Printf("-->>%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
}

func pollQueues(awsClient *aws.SqsClient, queueList []app.Queue) {

	m := app.Message{}
	m.Type = "message"
	m.Channel = slackChannel
	m.User = "@here"
	for {
		getQueueDepths(queueList, awsClient)
		outputMessage(queueList)
		<-time.After(10 * time.Minute)
	}
}

func getQueueDepths(queueList []app.Queue, awsClient *aws.SqsClient) {
	var err error
	for x := 0; x < len(queueList); x++ {
		queueList[x].LastDepth, err = strconv.Atoi(awsClient.GetSQSQueueDepth(queueList[x].Name))
		if err != nil {
			fmt.Println("Error:", err.Error())
		}
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

	// Sort the list
	sort.Slice(queueList, func(i, j int) bool {
		return queueList[i].Name < queueList[j].Name
	})

	fmt.Printf("%+v", queueList)

	awsClient := aws.CreateSqsClient(awsServer, awsRegion, awsClientID)

	go pollQueues(awsClient, queueList)
	slackListener(awsClient, queueList)
}

func slackListener(awsClient *aws.SqsClient, queueList []app.Queue) {
	ws, id := lslack.SlackConnect(os.Getenv("SLACK_TOKEN"))
	for {
		// read each incoming message
		m, muid, err := lslack.GetMessage(apiToken, ws)
		if err != nil {
			log.Println(err)
		}

		// see if we're mentioned
		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			msg := strings.TrimSpace(strings.TrimPrefix(m.Text, "<@"+id+">"))
			cmdParts := strings.Split(msg, " ")
			if cmdParts[0] == "help" {
				m.Text = "*QueueDepthBot Commands:* \n" +
					"\t*help* - This message\n" +
					"\t*list* - List all the queues being monitored\n" +
					"\t*add <queue>:<alert value> ...*  -  Add queue(s) to be monitored with their alert threshold\n" +
					"\t*del <queue> ...* - Remove the queue(s) from the monitoring list\n" +
					"\t*pause <queue> ...* - Pause the monitoring of the queue(s)\n" +
					"\t*unpause <queue> ...* - Unpause the monitoring of the queue(s)\n" +
					"\t*report* - Run the queue depth on all the queues and output the report"
				lslack.PostMessage(ws, m)
			} else if cmdParts[0] == "list" {
				m.Text = "*Queues being monitored:* \n```"
				for x := 0; x < len(queueList); x = x + 2 {
					tmp := fmt.Sprintf(" %s(%d)", queueList[x].Name, queueList[x].MaxDepth)
					if queueList[x].Paused {
						tmp = fmt.Sprintf("*%s(%d)", queueList[x].Name, queueList[x].MaxDepth)
					}
					tmp = fmt.Sprintf("%-55s", tmp)
					m.Text = m.Text + tmp
					if (x + 1) < len(queueList) {
						tmp = " "
						if queueList[x+1].Paused {
							tmp = "*"
						}
						tmp = fmt.Sprintf("%s%s(%d)\n", tmp, queueList[x+1].Name, queueList[x+1].MaxDepth)
						m.Text = m.Text + tmp
					} else {
						m.Text = m.Text + "\n"
					}
				}
				m.Text = m.Text + "```"
				lslack.PostMessage(ws, m)
			} else if cmdParts[0] == "add" {
				for x := 1; x < len(cmdParts); x++ {
					queue := cmdParts[x]
					alertSize := 1
					if strings.Contains(cmdParts[x], ":") {
						parts := strings.Split(cmdParts[x], ":")
						queue = parts[0]
						alertSize, _ = strconv.Atoi(parts[1])
					}
					queueList = append(queueList, app.Queue{
						MaxDepth: alertSize,
						Name:     queue,
					})
				}
				m.Text = "*Add Complete*"
				lslack.PostMessage(ws, m)
			} else if cmdParts[0] == "del" {
				for x := 1; x < len(cmdParts); x++ {
					for i := 0; i < len(queueList); i++ {
						if queueList[i].Name == cmdParts[x] {
							queueList = append(queueList[:i], queueList[i+1:]...)
						}
					}
				}
				m.Text = "*Del Complete*"
				lslack.PostMessage(ws, m)
			} else if cmdParts[0] == "pause" {
				for x := 1; x < len(cmdParts); x++ {
					for i := 0; i < len(queueList); i++ {
						if queueList[i].Name == cmdParts[x] {
							queueList[i].Paused = true
						}
					}
				}
				m.Text = "*Pause Complete*"
				lslack.PostMessage(ws, m)
			} else if cmdParts[0] == "unpause" {
				for x := 1; x < len(cmdParts); x++ {
					for i := 0; i < len(queueList); i++ {
						if queueList[i].Name == cmdParts[x] {
							queueList[i].Paused = false
						}
					}
				}
				m.Text = "*Unpause Complete*"
				lslack.PostMessage(ws, m)
			} else if cmdParts[0] == "report" {
				getQueueDepths(queueList, awsClient)
				outputMessage(queueList)
			} else {
				m.Text = "<@" + muid + "> Sorry did not understand what you wanted try to ask me for 'help'"
				lslack.PostMessage(ws, m)
			}
		}
	}
}
