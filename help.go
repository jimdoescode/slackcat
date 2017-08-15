package main

import (
	"bytes"
	"fmt"
	"github.com/nlopes/slack"
	"reflect"
	"text/tabwriter"
)

type HelpCommand struct {
	rtm  *slack.RTM
	cmds []SlackCatCommand
}

func (c *HelpCommand) Matches(msg *slack.Msg) (bool, bool) {
	return msg.Text == "?help", false
}

func (c *HelpCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	buf := bytes.NewBufferString("Here are all my known commands...\n```")
	w := tabwriter.NewWriter(buf, 4, 0, 1, ' ', tabwriter.AlignRight)
	for _, cmd := range c.cmds {
		r := reflect.TypeOf(cmd).Elem()
		fmt.Fprintf(w, "%s:\t %s\n", r.Name(), cmd.GetSyntax())
	}
	fmt.Fprint(w, "```")
	w.Flush()

	return c.rtm.NewOutgoingMessage(buf.String(), msg.Channel), nil
}

func (c *HelpCommand) GetSyntax() string {
	return "?help"
}

func (c *HelpCommand) Close() {
}

func NewHelpCommand(rtm *slack.RTM, cmds []SlackCatCommand) *HelpCommand {
	return &HelpCommand{rtm, cmds}
}
