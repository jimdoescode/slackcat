package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"os"
)

func main() {

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: slackcat <slack-bot-token> <slack-user-id>\n")
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "slack-cat: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	client := slack.New(os.Args[1])
	_, _, adminChan, err := client.OpenIMChannel(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not establish admin PM channel")
		os.Exit(1)
	}

	//TODO: Add new webhooks to this slice
	callbacks := []SlackCatCallback{
		NewSonarrCallback(client, adminChan), //Sonarr Responses will just go to the admin
	}

	defer func(callbacks []SlackCatCallback) {
		for _, callback := range callbacks {
			callback.Close()
		}
	}(callbacks)

	rtm := client.NewRTM()
	defer rtm.Disconnect()
	go rtm.ManageConnection()

	//TODO: Add commands to this slice
	cmds := []SlackCatCommand{
		NewPlusCommand(rtm),
		NewPlusDenominationCommand(rtm),
		NewGiphyCommand(rtm),
		NewHaltCommand(rtm),
		NewUpdateCommand(rtm),
		//Learn command should match everything so keep it last
		NewLearnCommand(rtm),
	}

	defer func(cmds []SlackCatCommand) {
		for _, cmd := range cmds {
			cmd.Close()
		}
	}(cmds)

	disconnect := false

	for msg := range rtm.IncomingEvents {
		if disconnect {
			break
		}

		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			for _, cmd := range cmds {
				if cmd.Matches(&ev.Msg) {
					out, err := cmd.Execute(&ev.Msg)
					if err != nil {
						fmt.Printf("Command Error: %v\n", err)
					}

					if out != nil {
						rtm.SendMessage(out)
					}
					break
				}
			}

		case *slack.DisconnectedEvent:
			disconnect = ev.Intentional
			break

		}
	}
}

type SlackCatCommand interface {
	Matches(msg *slack.Msg) bool
	Execute(msg *slack.Msg) (*slack.OutgoingMessage, error)
	GetSyntax() string
	Close()
}

type SlackCatCallback interface {
	Handle(blob []byte) error
	Close()
}
