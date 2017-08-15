package main

import (
	"github.com/nlopes/slack"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type GifCommand struct {
	rtm    *slack.RTM
	client *http.Client
	match  *regexp.Regexp
}

func (c *GifCommand) Matches(msg *slack.Msg) (bool, bool) {
	return strings.HasPrefix(msg.Text, "?gif "), false
}

func (c *GifCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	txt := strings.SplitN(msg.Text, " ", 2)
	q := url.QueryEscape(
		strings.ToLower(parseUsernamesAndChannels(&c.rtm.Client, txt[1])),
	)

	searchUrl := "https://www.google.com/search?source=lnms&tbm=isch&tbs=itp:animated,ift:gif&q=" + q

	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.71 Safari/537.36")

	resp, err := c.client.Do(req)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	resp.Body.Close()
	found := c.match.FindStringSubmatch(string(body[:]))

	if found == nil {
		out := c.rtm.NewOutgoingMessage("I got nothing for that.", msg.Channel)
		return out, nil
	}

	return c.rtm.NewOutgoingMessage(found[1], msg.Channel), nil
}

func (c *GifCommand) GetSyntax() string {
	return "?gif <search>"
}

func (c *GifCommand) GetDescription() string {
	return "Uses google image search to find a gif matching the search phrase"
}

func (c *GifCommand) Close() {
}

func NewGifCommand(rtm *slack.RTM) *GifCommand {
	return &GifCommand{
		rtm,
		&http.Client{},
		regexp.MustCompile(`"ou":"(.*?)"`),
	}
}
