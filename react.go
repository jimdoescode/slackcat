package main

import (
	"database/sql"
	"fmt"
	"github.com/nlopes/slack"
	"regexp"
	"strings"
)

type ReactCommand struct {
	rtm *slack.RTM
	exp *regexp.Regexp
	ins *sql.Stmt
	del *sql.Stmt
	sel *sql.Stmt
}

func (c *ReactCommand) Matches(msg *slack.Msg) (bool, bool) {
	if c.exp.MatchString(msg.Text) {
		return true, false
	}

	txt := strings.ToLower(strings.TrimSpace(msg.Text))
	if len(txt) < 1 {
		return false, false
	}

	var val string
	err := c.sel.QueryRow(txt).Scan(&val)
	return err == nil, true
}

func (c *ReactCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	msgRef := slack.NewRefToMessage(msg.Channel, msg.Timestamp)

	if c.exp.MatchString(msg.Text) {
		vars := c.exp.FindStringSubmatch(msg.Text)
		dbCmd := c.ins
		out := c.rtm.NewOutgoingMessage("Got it.", msg.Channel)
		if vars[1] == "unreact" {
			dbCmd = c.del
			out.Text = fmt.Sprintf("Removed :%s: reaction", vars[2])
		}

		_, err := dbCmd.Exec(vars[3], vars[2])
		if err != nil {
			return nil, err
		}

		return out, nil
	}

	txt := strings.ToLower(msg.Text)
	if len(txt) < 1 {
		return nil, nil
	}

	rows, err := c.sel.Query(txt)
	if err != nil {
		rows.Close()
		return nil, err
	}

	for rows.Next() {
		var emoji string
		if err := rows.Scan(&emoji); err != nil {
			continue
		}

		c.rtm.AddReaction(emoji, msgRef)
	}

	rows.Close()

	return nil, nil
}

func (c *ReactCommand) GetSyntax() string {
	return "?(un)react <emoji> to <string>"
}

func (c *ReactCommand) Close() {
	c.sel.Close()
	c.del.Close()
	c.ins.Close()
}

func NewReactCommand(rtm *slack.RTM, db *sql.DB) *ReactCommand {
	exp := regexp.MustCompile(`^(?i)\?(react|unreact) :(\w+?): to (.+?)$`)

	db.Exec("CREATE TABLE reactions (target TEXT NOT NULL, emoji TEXT NOT NULL)")
	db.Exec("CREATE INDEX target_idx IF NOT EXISTS ON reactions (target)")
	db.Exec("CREATE INDEX target_emoji_idx IF NOT EXISTS ON reactions (target, emoji)")

	ins, err := db.Prepare("INSERT INTO reactions(target, emoji) VALUES(?,?)")
	if err != nil {
		fmt.Printf("error preparing reactions insert: %v\n", err)
		return nil
	}

	del, err := db.Prepare("DELETE from reactions WHERE target=? AND emoji=?")
	if err != nil {
		fmt.Printf("error preparing reactions delete: %v\n", err)
		return nil
	}

	sel, err := db.Prepare("SELECT emoji FROM reactions WHERE target=?")
	if err != nil {
		fmt.Printf("error preparing reactions select: %v\n", err)
		return nil
	}

	return &ReactCommand{rtm, exp, ins, del, sel}
}
