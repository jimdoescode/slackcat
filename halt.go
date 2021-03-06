package main

import (
	"github.com/nlopes/slack"
)

type HaltCommand struct {
	rtm *slack.RTM
}

func (c *HaltCommand) Matches(msg *slack.Msg) (bool, bool) {
	return msg.Text == "?halt", false
}

func (c *HaltCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	status := c.rtm.NewOutgoingMessage("Brb...", msg.Channel)
	c.rtm.SendMessage(status)
	c.rtm.Disconnect()

	return nil, nil
}

func (c *HaltCommand) GetSyntax() string {
	return "?halt"
}

func (c *HaltCommand) GetDescription() string {
	return "Shut down slack cat. It should automatically restart"
}

func (c *HaltCommand) Close() {
}

func NewHaltCommand(rtm *slack.RTM) *HaltCommand {
	return &HaltCommand{rtm}
}
