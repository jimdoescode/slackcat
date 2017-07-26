package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"os"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: slackcat slack-bot-token\n")
		os.Exit(1)
	}

	client := slack.New(os.Args[1])
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	rtm := client.NewRTM()
	defer rtm.Disconnect()
	go rtm.ManageConnection()

	//TODO: Add commands to this slice
	cmds := []SlackCatCommand{
		NewPlusCommand(rtm),
		NewPlusDenominationCommand(rtm),
		NewGiphyCommand(rtm),
		NewLearnCommand(rtm),
	}

	defer func(cmds []SlackCatCommand) {
		for _, cmd := range cmds {
			cmd.Close()
		}
	}(cmds)

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			for _, cmd := range cmds {
				if cmd.Matches(&ev.Msg) {
					out, _ := cmd.Execute(&ev.Msg)
					if out != nil {
						rtm.SendMessage(out)
						break
					}
				}
			}
		default:

		}
	}
}

type SlackCatCommand interface {
	Matches(msg *slack.Msg) bool
	Execute(msg *slack.Msg) (*slack.OutgoingMessage, error)
	GetSyntax() string
	Close()
}
