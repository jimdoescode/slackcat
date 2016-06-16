package main

import (
	"encoding/json"
	"fmt"
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
	cli    *http.Client
	search *url.URL
	key    string
}

func (c *GiphyCommand) Execute(msg *SlackMessage) (*SlackMessage, error) {
	txt := strings.SplitN(msg.Text, " ", 2)
	token := strings.ToLower(txt[0][1:])

	if token != "giphy" {
		return nil, nil
	}

	if len(txt) < 2 {
		msg.Text = c.GetSyntax()
		return msg, nil
	}

	q := c.search.Query()
	q.Set("api_key", "dc6zaTOxFJmzC")
	q.Set("q", txt[1])
	q.Set("limit", "100")
	c.search.RawQuery = q.Encode()

	resp, err := c.cli.Get(c.search.String())
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respObj giphyResp
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return nil, err
	}

	if respObj.Meta.Status != 200 {
		return nil, fmt.Errorf("Giphy error: %s", respObj.Meta.Error)
	}

	if len(respObj.Data) == 0 {
		msg.Text = "Giphy don't know"
	} else {
		rand.Seed(time.Now().Unix())
		randData := respObj.Data[rand.Intn(len(respObj.Data))]
		msg.Text = randData.Images["downsized"].Url
	}

	return msg, nil
}

func (c *GiphyCommand) GetSyntax() string {
	return "Syntax: ?giphy <search>"
}

func (c *GiphyCommand) Close() {

}

func NewGiphyCommand() *GiphyCommand {
	search := &url.URL{
		Scheme: "http",
		Host:   "api.giphy.com",
		Path:   "v1/gifs/search",
	}

	return &GiphyCommand{
		&http.Client{},
		search,
		"dc6zaTOxFJmzC",
	}
}
