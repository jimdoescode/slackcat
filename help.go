package main

import (
	"bytes"
	"fmt"
	"github.com/nlopes/slack"
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
	f := "%s\n\t%s\n\n"

	//Add the help command syntax
	fmt.Fprintf(w, f, c.GetSyntax(), c.GetDescription())

	for _, cmd := range c.cmds {
		fmt.Fprintf(w, f, cmd.GetSyntax(), cmd.GetDescription())
	}

	fmt.Fprint(w, "```")
	w.Flush()

	return c.rtm.NewOutgoingMessage(buf.String(), msg.Channel), nil
}

func (c *HelpCommand) GetSyntax() string {
	return "?help"
}

func (c *HelpCommand) GetDescription() string {
	return "Display this help message"
}

func (c *HelpCommand) Close() {
}

func NewHelpCommand(rtm *slack.RTM, cmds []SlackCatCommand) *HelpCommand {
	return &HelpCommand{rtm, cmds}
}
