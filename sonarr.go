package main

import (
	"github.com/nlopes/slack"
)

type SonarrCallback struct {
	client   *slack.Client
	respChan string
}

func (c *SonarrCallback) Handle(blob []byte) error {
	_, _, err := c.client.PostMessage(c.respChan, "ping", slack.NewPostMessageParameters())
	return err
}

func (c *SonarrCallback) Close() {

}

func NewSonarrCallback(client *slack.Client, respChan string) *SonarrCallback {
	return &SonarrCallback{client, respChan}
}
