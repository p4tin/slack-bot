package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"

	"golang.org/x/net/websocket"
)

// These two structures represent the response of the Slack API rtm.start.
// Only some fields are included. The rest are ignored by json.Unmarshal.

type responseRtmStart struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	Url   string       `json:"url"`
	Self  responseSelf `json:"self"`
}

type responseSelf struct {
	Id string `json:"id"`
}

// slackStart does a rtm.start, and returns a websocket URL and user ID. The
// websocket URL can be used to initiate an RTM session.
func SlackStart(apiToken string) (wsurl, id string, err error) {
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", apiToken)
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
	var respObj responseRtmStart
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

type Message struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
	User    string `json: "user"`
}

func GetMessage(apiToken string, ws *websocket.Conn) (m Message, uid string, err error) {
	err = websocket.JSON.Receive(ws, &m)
	username, _ := GetUserName(apiToken, m.User)
	uid = m.User
	m.User = username
	return
}

var counter uint64

func PostMessage(ws *websocket.Conn, m Message) error {
	m.Id = atomic.AddUint64(&counter, 1)
	return websocket.JSON.Send(ws, m)
}

// Starts a websocket-based Real Time API session and return the websocket
// and the ID of the (bot-)user whom the token belongs to.
func SlackConnect(apiToken string) (*websocket.Conn, string) {
	wsurl, id, err := SlackStart(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	ws, err := websocket.Dial(wsurl, "", "https://api.slack.com/")
	if err != nil {
		log.Fatal(err)
	}

	return ws, id
}

type UserInfo struct {
	Ok   bool `json:"ok"`
	User struct {
		Color             string `json:"color"`
		Deleted           bool   `json:"deleted"`
		ID                string `json:"id"`
		IsAdmin           bool   `json:"is_admin"`
		IsBot             bool   `json:"is_bot"`
		IsOwner           bool   `json:"is_owner"`
		IsPrimaryOwner    bool   `json:"is_primary_owner"`
		IsRestricted      bool   `json:"is_restricted"`
		IsUltraRestricted bool   `json:"is_ultra_restricted"`
		Name              string `json:"name"`
		Profile           struct {
			Email              string `json:"email"`
			FirstName          string `json:"first_name"`
			Image192           string `json:"image_192"`
			Image24            string `json:"image_24"`
			Image32            string `json:"image_32"`
			Image48            string `json:"image_48"`
			Image72            string `json:"image_72"`
			ImageOriginal      string `json:"image_original"`
			LastName           string `json:"last_name"`
			Phone              string `json:"phone"`
			RealName           string `json:"real_name"`
			RealNameNormalized string `json:"real_name_normalized"`
			Skype              string `json:"skype"`
			Title              string `json:"title"`
		} `json:"profile"`
		RealName string      `json:"real_name"`
		Status   interface{} `json:"status"`
		TeamID   string      `json:"team_id"`
		Tz       string      `json:"tz"`
		TzLabel  string      `json:"tz_label"`
		TzOffset int         `json:"tz_offset"`
	} `json:"user"`
}

func GetUserName(apiToken string, userId string) (username string, success bool) {
	url := "https://slack.com/api/users.info?token=" + apiToken + "&user=" + userId + "&pretty=1"
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
	var user UserInfo
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
