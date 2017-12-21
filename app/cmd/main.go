package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"slack-bot/app"
	"slack-bot/app/clients/slack"
)

var apiToken string

func pollQueues(slackClient *slack.SlackClient) {
	m := app.Message{}
	m.Text = "Hello World!!!"
	m.User = "@here"
	for {
		<-time.After(5 * time.Second)
		err := slackClient.PostMessage(m)
		if err != nil {
			fmt.Printf("Error sending Hello World Message: %s\n", err.Error())
		}
	}
}

func main() {
	apiToken = "xoxb-21530270550-46NmEGWv6F4VIbY0WtmJvUaR"
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

			if len(parts) == 3 && parts[1] == "stock" {
				// looks good, get the quote and reply with the result
				go func(m app.Message) {
					m.Text = getQuote(m.User, parts[2])
					slackClient.PostMessage(m)
				}(m)
				// NOTE: the Message object is copied, this is intentional
			} else if len(parts) > 2 && parts[1] == "save" {
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

// Get the quote via Yahoo. You should replace this method to something
// relevant to your team!
func getQuote(user string, sym string) string {
	sym = strings.ToUpper(sym)
	url := fmt.Sprintf("http://download.finance.yahoo.com/d/quotes.csv?s=%s&f=nsl1op&e=.csv", sym)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	rows, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if len(rows) >= 1 && len(rows[0]) == 5 {
		return fmt.Sprintf("@%s - %s (%s) is trading at $%s", user, rows[0][0], rows[0][1], rows[0][2])
	}
	return fmt.Sprintf("unknown response format (symbol was \"%s\")", sym)
}
