package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {

	//TODO: Add commands to this slice
	cmds := []SlackCatCommand{
		NewPlusCommand(),
		NewPlusDenominationCommand(),
		NewGiphyCommand(),
		NewLearnCommand(),
	}

	defer func(cmds []SlackCatCommand) {
		for _, cmd := range cmds {
			cmd.Close()
		}
	}(cmds)

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: slackcat slack-bot-token\n")
		os.Exit(1)
	}

	bot, err := NewSlackBot(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	for {
		msg, err := bot.GetMessage()

		if err != nil {
			fmt.Printf("error getting or sending a message: %v\n", err)
			continue
		}

		//If there isn't a message or it doesn't start with a question mark
		if msg == nil || !strings.HasPrefix(msg.Text, "?") || len(msg.Text) < 2 {
			continue
		}

		msg.Text = bot.DecodeText(msg.Text)

		//TODO If they type help or something then we want to handle that here and pass that to the commands to determine which command they might want.

		go findAndExecute(msg, cmds, bot)
	}
}

func findAndExecuteHelp(msg *SlackMessage, cmds []SlackCatCommand, bot *SlackBot) {

}

func findAndExecute(msg *SlackMessage, cmds []SlackCatCommand, bot *SlackBot) {
	for _, cmd := range cmds {
		resp, err := cmd.Execute(msg)
		if err != nil {
			fmt.Printf("error executing command: %v\n", err)
			break
		}

		if resp == nil {
			continue
		}

		resp.Channel = msg.Channel
		resp.Id = msg.Id
		resp.Type = msg.Type

		if err != nil {
			fmt.Printf("error sending message: %v\n", err)
		}

		bot.PostMessage(msg)

		break
	}
}

type SlackCatCommand interface {
	Execute(msg *SlackMessage) (*SlackMessage, error)
	GetSyntax() string
	Close()
}
