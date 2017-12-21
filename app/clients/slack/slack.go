package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
	"slack-bot/app"

	"golang.org/x/net/websocket"
)

type SlackClient struct {
	apiToken string
	ws       *websocket.Conn
}

func CreateSlackClient(token string) *SlackClient {
	return &SlackClient{
		apiToken: token,
	}
}

// Starts a websocket-based Real Time API session and return the websocket
// and the ID of the (bot-)user whom the token belongs to.
func (sc *SlackClient) SlackConnect() string {
	wsurl, id, err := sc.SlackStart()
	if err != nil {
		log.Fatal(err)
	}

	ws, err := websocket.Dial(wsurl, "", "https://api.slack.com/")
	if err != nil {
		log.Fatal(err)
	}
	sc.ws = ws
	return id
}

func (sc *SlackClient) SlackStart() (wsurl, id string, err error) {
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", sc.apiToken)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	var respObj app.ResponseRtmStart
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return
	}

	if !respObj.Ok {
		err = fmt.Errorf("Slack error: %s", respObj.Error)
		return
	}

	wsurl = respObj.Url
	id = respObj.Self.Id
	return
}

// These are the messages read off and written into the websocket. Since this
// struct serves as both read and write, we include the "Id" field which is
// required only for writing.

func (sc *SlackClient) GetMessage() (app.Message, error) {
	m := app.Message{}
	err := websocket.JSON.Receive(sc.ws, &m)
	if err != nil {
		return m, err
	}
	username, succ := sc.GetUserName(m.User)
	if succ {
		m.User = username
	} else {
		m.User = "Unknown"
	}
	return m, nil
}

var counter uint64

func (sc *SlackClient) PostMessage(m app.Message) error {
	m.Id = atomic.AddUint64(&counter, 1)
	return websocket.JSON.Send(sc.ws, m)
}

func (sc *SlackClient) GetUserName(userId string) (username string, success bool) {
	url := "https://slack.com/api/users.info?token=" + sc.apiToken + "&user=" + userId + "&pretty=1"
	resp, err := http.Get(url)
	if err != nil {
		return userId, false
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return userId, false
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return userId, false
	}
	var user app.UserInfo
	err = json.Unmarshal(body, &user)
	if err != nil {
		return userId, false
	}

	if !user.Ok {
		err = fmt.Errorf("Slack error: Error retrieving User Info for user: %s\n", userId)
		return userId, false
	}

	return user.User.Name, true
}
