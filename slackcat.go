package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nlopes/slack"
	"log"
	"os"
	"path/filepath"
)

func main() {

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: slackcat <slack-bot-token> <slack-user-id>\n")
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "slack-cat: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not determine executable location")
		os.Exit(1)
	}

	loc := filepath.Dir(exe)
	db, err := sql.Open("sqlite3", filepath.Join(loc, "slackcat.db"))
	defer db.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open database connection")
		os.Exit(1)
	}

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
		NewReactCommand(rtm, db),
		NewPlusCommand(rtm, db),
		NewPlusDenominationCommand(rtm, db),
		NewGifCommand(rtm),
		NewGiphyCommand(rtm),
		NewHaltCommand(rtm),
		NewUpdateCommand(rtm),
		//Learn command should match everything so keep it last
		NewLearnCommand(rtm, db),
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
				match, cont := cmd.Matches(&ev.Msg)
				if match {
					out, err := cmd.Execute(&ev.Msg)
					if err != nil {
						fmt.Printf("Command Error: %v\n", err)
					}

					if out != nil {
						rtm.SendMessage(out)
					}

					if !cont {
						break
					}
				}
			}

		case *slack.DisconnectedEvent:
			disconnect = ev.Intentional
			break

		}
	}
}

type SlackCatCommand interface {
	Matches(msg *slack.Msg) (bool, bool)
	Execute(msg *slack.Msg) (*slack.OutgoingMessage, error)
	GetSyntax() string
	Close()
}

type SlackCatCallback interface {
	Handle(blob []byte) error
	Close()
}
