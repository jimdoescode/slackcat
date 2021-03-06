package main

// NOTE
// This command expects that the binary is being run in the
// same directory that the code and git repo exist in.

import (
	"bytes"
	"fmt"
	"github.com/nlopes/slack"
	"gopkg.in/src-d/go-git.v4"
	"os"
	"os/exec"
	"path/filepath"
)

type UpdateCommand struct {
	rtm *slack.RTM
}

func (c *UpdateCommand) Matches(msg *slack.Msg) (bool, bool) {
	return msg.Text == "?update", false
}

func (c *UpdateCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	status := c.rtm.NewOutgoingMessage("Updating repo...", msg.Channel)
	c.rtm.SendMessage(status)

	exe, err := os.Executable()
	if err != nil {
		status = c.rtm.NewOutgoingMessage("Could not determine repo location. Halting update.", msg.Channel)
		return status, err
	}

	root := filepath.Dir(exe)
	repo, err := git.PlainOpen(root)
	if err != nil {
		status = c.rtm.NewOutgoingMessage("Error opening repository. Halting update.", msg.Channel)
		return status, err
	}

	err = repo.Pull(&git.PullOptions{})
	if err == git.NoErrAlreadyUpToDate || err == nil {
		status = c.rtm.NewOutgoingMessage("Repo updated. Recompiling...", msg.Channel)
		c.rtm.SendMessage(status)
	} else {
		status = c.rtm.NewOutgoingMessage("Error pulling repository. Halting update.", msg.Channel)
		return status, err
	}

	var stderr bytes.Buffer
	cmd := exec.Command("go", "build")
	cmd.Stderr = &stderr
	cmd.Dir = root
	err = cmd.Run()

	if err != nil {
		status = c.rtm.NewOutgoingMessage(
			fmt.Sprintf("Error recompiling (%s). Halting update.", stderr.String()),
			msg.Channel,
		)
		return status, err
	}

	status = c.rtm.NewOutgoingMessage("Recompile done. Brb...", msg.Channel)
	c.rtm.SendMessage(status)

	c.rtm.Disconnect()

	return nil, nil
}

func (c *UpdateCommand) GetSyntax() string {
	return "?update"
}

func (c *UpdateCommand) GetDescription() string {
	return "Make slack cat update its repo and recompile."
}

func (c *UpdateCommand) Close() {
}

func NewUpdateCommand(rtm *slack.RTM) *UpdateCommand {
	return &UpdateCommand{rtm}
}
