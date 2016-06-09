package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
)

type responseRtmStart struct {
	Ok       bool                 `json:"ok"`
	Error    string               `json:"error"`
	Url      string               `json:"url"`
	Self     responseRtmSelf      `json:"self"`
	Channels []responseRtmChannel `json:"channels"`
	Users    []responseRtmUser    `json:"users"`
}

type responseRtmSelf struct {
	Id string `json:"id"`
}

type responseRtmChannel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type responseRtmUser struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewSlackBot(token string) (*SlackBot, error) {
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var respObj responseRtmStart
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return nil, err
	}

	if !respObj.Ok {
		return nil, fmt.Errorf("Slack error: %s", respObj.Error)
	}

	ws, err := websocket.Dial(respObj.Url, "", "https://api.slack.com/")
	if err != nil {
		return nil, err
	}

	//Move users and channels into maps keyed by id
	users := map[string]string{}
	for _, rtmuser := range respObj.Users {
		users[rtmuser.Id] = rtmuser.Name
	}

	channels := map[string]string{}
	for _, rtmchan := range respObj.Channels {
		channels[rtmchan.Id] = rtmchan.Name
	}

	return &SlackBot{respObj.Self.Id, ws, 0, users, channels}, nil
}

type SlackBot struct {
	id       string
	conn     *websocket.Conn
	counter  uint64
	users    map[string]string
	channels map[string]string
}

type Message struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func (b *SlackBot) GetMessage() (*Message, error) {
	var msg Message
	err := websocket.JSON.Receive(b.conn, &msg)
	if err != nil {
		return nil, err
	}

	if msg.Type == "message" {
		return &msg, nil
	}

	return nil, nil
}

func (b *SlackBot) ParseText(txt string) string {

	re := regexp.MustCompile("<([@#!])?([^>|]+)(?:\\|([^>]+))?>")
	tokens := re.FindAllStringSubmatch(txt, -1)
	special := map[string]bool{
		"channel":  true,
		"group":    true,
		"everyone": true,
		"here":     true,
	}

	for _, token := range tokens {
		//This could be a user so look up their name
		if token[1] == "@" {
			user := b.users[token[2]]
			if user == "" {
				user = token[2]
			}
			txt = strings.Replace(txt, token[0], fmt.Sprintf("@%s", user), 1)
			//This could be a channel so look it up
		} else if token[1] == "#" {
			chann := b.channels[token[2]]
			if chann == "" {
				chann = token[2]
			}
			txt = strings.Replace(txt, token[0], fmt.Sprintf("#%s", chann), 1)
			//This could be a special command so look it up
		} else if token[1] == "!" {
			if special[token[2]] {
				txt = strings.Replace(txt, token[0], fmt.Sprintf("@%s", token[2]), 1)
			}
			//This is a link so strip out email stuff
		} else {
			link := strings.Replace(token[2], "mailto:", "", -1)
			//If this link has a label add it
			if len(token) > 2 && len(token[3]) > 0 {
				link = fmt.Sprintf("%s (%s)", token[3], link)
			}
			txt = strings.Replace(txt, token[0], link, 1)
		}
	}

	return txt
}

func (b *SlackBot) PostMessage(msg *Message) error {
	msg.Id = atomic.AddUint64(&b.counter, 1)
	return websocket.JSON.Send(b.conn, *msg)
}
