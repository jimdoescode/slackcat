package main

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type giphyResp struct {
	Data []struct {
		Type   string `json:"type"`
		Id     string `json:"id"`
		Images map[string]struct {
			Url    string
			Width  string
			Height string
		} `json:"images"`
	} `json:"data"`
	Meta struct {
		Status int    `json:"status"`
		Error  string `json:"msg"`
	} `json:"meta"`
}

type GiphyCommand struct {
	rtm    *slack.RTM
	cli    *http.Client
	search *url.URL
	key    string
}

func (c *GiphyCommand) Matches(msg *slack.Msg) (bool, bool) {
	return strings.HasPrefix(msg.Text, "?giphy "), false
}

func (c *GiphyCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	txt := strings.SplitN(msg.Text, " ", 2)

	if len(txt) < 2 {
		return nil, fmt.Errorf("Invalid Syntax")
	}

	q := c.search.Query()
	q.Set("api_key", "dc6zaTOxFJmzC")
	q.Set("q", txt[1])
	q.Set("limit", "100")
	c.search.RawQuery = q.Encode()

	resp, err := c.cli.Get(c.search.String())

	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	var respObj giphyResp
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	if respObj.Meta.Status != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("Giphy error: %s", respObj.Meta.Error)
	}

	out := c.rtm.NewOutgoingMessage("Giphy don't know", msg.Channel)
	if len(respObj.Data) > 0 {
		rand.Seed(time.Now().Unix())
		randData := respObj.Data[rand.Intn(len(respObj.Data))]
		out.Text = randData.Images["downsized"].Url
	}

	return out, nil
}

func (c *GiphyCommand) GetSyntax() string {
	return "?giphy <search>"
}

func (c *GiphyCommand) Close() {

}

func NewGiphyCommand(rtm *slack.RTM) *GiphyCommand {
	search := &url.URL{
		Scheme: "http",
		Host:   "api.giphy.com",
		Path:   "v1/gifs/search",
	}

	return &GiphyCommand{
		rtm,
		&http.Client{},
		search,
		"dc6zaTOxFJmzC",
	}
}
