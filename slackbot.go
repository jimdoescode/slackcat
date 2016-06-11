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

type rtmStart struct {
	Ok       bool      `json:"ok"`
	Error    string    `json:"error"`
	Url      string    `json:"url"`
	Self     self      `json:"self"`
	Channels []channel `json:"channels"`
	Users    []user    `json:"users"`
}

type self struct {
	Id string `json:"id"`
}

type channel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type user struct {
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

	var respObj rtmStart
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
	return &SlackBot{respObj.Self.Id, ws, 0, respObj.Users, respObj.Channels}, nil
}

type SlackBot struct {
	id       string
	conn     *websocket.Conn
	counter  uint64
	users    []user
	channels []channel
}

type SlackMessage struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
	User    string `json:"user"`
}

func (b *SlackBot) GetUserName(id string) string {
	for _, u := range b.users {
		if id == u.Id {
			return u.Name
		}
	}
	return ""
}

func (b *SlackBot) DecodeText(txt string) string {
	special := map[string]bool{
		"channel":  true,
		"group":    true,
		"everyone": true,
		"here":     true,
	}

	users := map[string]string{}
	for _, u := range b.users {
		users[u.Id] = u.Name
	}

	channels := map[string]string{}
	for _, c := range b.channels {
		channels[c.Id] = c.Name
	}

	re := regexp.MustCompile("<([@#!])?([^>|]+)(?:\\|([^>]+))?>")

	for _, token := range re.FindAllStringSubmatch(txt, -1) {
		//This could be a user so look up their name
		if token[1] == "@" {
			user := users[token[2]]
			if user != "" {
				txt = strings.Replace(txt, token[0], fmt.Sprintf("@%s", user), 1)
			}
			//This could be a channel so look it up
		} else if token[1] == "#" {
			chann := channels[token[2]]
			if chann != "" {
				txt = strings.Replace(txt, token[0], fmt.Sprintf("#%s", chann), 1)
			}
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

func (b *SlackBot) EncodeText(txt string) string {

	special := map[string]bool{
		"channel":  true,
		"group":    true,
		"everyone": true,
		"here":     true,
	}

	users := map[string]string{}
	for _, u := range b.users {
		users[u.Name] = u.Id
	}

	channels := map[string]string{}
	for _, c := range b.channels {
		channels[c.Name] = c.Id
	}

	re := regexp.MustCompile("([@#]){1}([^\\s]+)")

	for _, token := range re.FindAllStringSubmatch(txt, -1) {
		if token[1] == "@" {
			user := users[token[2]]
			if user != "" {
				user = fmt.Sprintf("<@%s|%s>", user, token[2])
				txt = strings.Replace(txt, token[0], user, 1)
			} else if special[token[2]] {
				txt = strings.Replace(txt, token[0], fmt.Sprintf("<!%s|%s>", token[2], token[2]), 1)
			}
		} else if token[1] == "#" {
			chann := channels[token[2]]
			if chann != "" {
				chann = fmt.Sprintf("<#%s|%s>", chann, token[2])
				txt = strings.Replace(txt, token[0], chann, 1)
			}
		}
	}

	return txt
}

func (b *SlackBot) GetMessage() (*SlackMessage, error) {
	var msg SlackMessage
	err := websocket.JSON.Receive(b.conn, &msg)
	if err != nil {
		return nil, err
	}

	//Only return message types that aren't from the slackbot
	if msg.Type == "message" && msg.User != b.id {
		msg.User = fmt.Sprintf("@%s", b.GetUserName(msg.User))
		return &msg, nil
	}

	return nil, nil
}

func (b *SlackBot) PostMessage(msg *SlackMessage) error {
	msg.Id = atomic.AddUint64(&b.counter, 1)
	msg.User = b.id
	msg.Text = b.EncodeText(msg.Text)
	return websocket.JSON.Send(b.conn, *msg)
}
