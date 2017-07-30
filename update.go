package main

// NOTE
// This command expects that the binary is being run in the
// same directory that the code and git repo exist in.

import (
	"github.com/nlopes/slack"
	"gopkg.in/src-d/go-git.v4"
	"os/exec"
)

type UpdateCommand struct {
	rtm *slack.RTM
}

func (c *UpdateCommand) Matches(msg *slack.Msg) bool {
	return msg.Text == "?update"
}

func (c *UpdateCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	status := c.rtm.NewOutgoingMessage("Updating repo...", msg.Channel)
	c.rtm.SendMessage(status)

	repo, err := git.PlainOpen("./")
	if err != nil {
		status = c.rtm.NewOutgoingMessage("Error opening repository. Halting update.", msg.Channel)
		return status, err
	}

	err = repo.Pull(&git.PullOptions{})
	if err == git.NoErrAlreadyUpToDate {
		status = c.rtm.NewOutgoingMessage("No changes detected. Halting update.", msg.Channel)
		return status, nil
	} else if err != nil {
		status = c.rtm.NewOutgoingMessage("Error pulling repository. Halting update.", msg.Channel)
		return status, err
	} else {
		status = c.rtm.NewOutgoingMessage("Repo updated. Recompiling...", msg.Channel)
		c.rtm.SendMessage(status)
	}

	cmd := exec.Command("go", "build")
	err = cmd.Run()

	if err != nil {
		status = c.rtm.NewOutgoingMessage("Error recompiling. Halting update.", msg.Channel)
		return status, err
	}

	status = c.rtm.NewOutgoingMessage("Recompile done. Disconnecting...", msg.Channel)
	c.rtm.SendMessage(status)

	c.rtm.Disconnect()

	return nil, nil
}

func (c *UpdateCommand) GetSyntax() string {
	return "?update"
}

func (c *UpdateCommand) Close() {
}

func NewUpdateCommand(rtm *slack.RTM) *UpdateCommand {
	return &UpdateCommand{rtm}
}
