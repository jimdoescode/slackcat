package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nlopes/slack"
	"regexp"
	"strings"
)

type LearnCommand struct {
	rtm *slack.RTM
	db  *sql.DB
	ins *sql.Stmt
	del *sql.Stmt
	sel *sql.Stmt
}

func (c *LearnCommand) Matches(msg *slack.Msg) bool {
	match := strings.HasPrefix(msg.Text, "?learn ") ||
		strings.HasPrefix(msg.Text, "?unlearn ")

	if !match {
		txt := strings.SplitN(msg.Text, " ", 2)
		if len(txt) < 1 || len(txt[0]) < 1 || txt[0][0] != '?' {
			return false
		}

		token := strings.ToLower(txt[0][1:])
		var val string
		err := c.sel.QueryRow(token).Scan(&val)
		match = err == nil
	}

	return match
}

func (c *LearnCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	txt := strings.SplitN(msg.Text, " ", 3)
	token := strings.ToLower(txt[0][1:])

	if token != "learn" && token != "unlearn" {
		var val string
		token = c.parseTarget(token)
		err := c.sel.QueryRow(token).Scan(&val)
		if err != nil {
			fmt.Printf("error searching db: %v\n", err)
			return nil, nil
		}

		out := c.rtm.NewOutgoingMessage(c.parseText(val), msg.Channel)
		return out, err
	}

	if len(txt) < 3 {
		out := c.rtm.NewOutgoingMessage(c.GetSyntax(), msg.Channel)
		return out, nil
	}

	target := c.parseTarget(txt[1])
	if token == "learn" {
		if target == token {
			out := c.rtm.NewOutgoingMessage("We must go deeper!", msg.Channel)
			return out, nil
		}

		_, err := c.ins.Exec(target, txt[2])
		out := c.rtm.NewOutgoingMessage(fmt.Sprintf("OK, learned %s", txt[1]), msg.Channel)
		return out, err
	}

	if token == "unlearn" {
		if target == token {
			out := c.rtm.NewOutgoingMessage("Don't incept me!", msg.Channel)
			return out, nil
		}

		_, err := c.del.Exec(target, txt[2])
		out := c.rtm.NewOutgoingMessage(fmt.Sprintf("Unlearned %s", txt[1]), msg.Channel)

		return out, err
	}

	return nil, nil
}

func (c *LearnCommand) GetSyntax() string {
	return "?(un)learn <target> <value>"
}

func (c *LearnCommand) Close() {
	c.sel.Close()
	c.del.Close()
	c.ins.Close()
	c.db.Close()
}

func (c *LearnCommand) parseTarget(txt string) string {
	userReg := regexp.MustCompile("^<@(\\w+)>$")
	chanReg := regexp.MustCompile("^<#(\\w+)\\|?(\\w*)>$")
	if userReg.MatchString(txt) {
		vars := userReg.FindStringSubmatch(txt)
		user, err := c.rtm.GetUserInfo(vars[1])
		if err == nil {
			txt = user.Name
		}
	} else if chanReg.MatchString(txt) {
		vars := chanReg.FindStringSubmatch(txt)
		ch, err := c.rtm.GetChannelInfo(vars[1])
		if err == nil {
			txt = ch.Name
		}
	}

	return strings.ToLower(txt)
}

func (c *LearnCommand) parseText(txt string) string {
	re := regexp.MustCompile("\\?([^\\s]+)")
	vars := re.FindAllStringSubmatch(txt, -1)

	for _, v := range vars {
		var val string
		err := c.sel.QueryRow(strings.ToLower(v[1])).Scan(&val)
		if err != nil {
			val = v[0]
		}

		txt = strings.Replace(txt, v[0], val, 1)
	}

	return txt
}

func NewLearnCommand(rtm *slack.RTM) *LearnCommand {
	db, err := sql.Open("sqlite3", "./slackcat.db")
	if err != nil {
		fmt.Printf("error creating learn command: %v\n", err)
		return nil
	}

	db.Exec("CREATE TABLE learns (target TEXT NOT NULL, value TEXT NOT NULL)")
	db.Exec("CREATE INDEX target_idx IF NOT EXISTS ON learns (target)")
	db.Exec("CREATE INDEX target_value_idx IF NOT EXISTS ON learns (target, value)")

	ins, err := db.Prepare("INSERT INTO learns(target, value) VALUES(?,?)")
	if err != nil {
		fmt.Printf("error preparing learn insert: %v\n", err)
		return nil
	}

	del, err := db.Prepare("DELETE from learns WHERE target=? AND value=?")
	if err != nil {
		fmt.Printf("error preparing learn delete: %v\n", err)
		return nil
	}

	sel, err := db.Prepare("SELECT value FROM learns WHERE target=? ORDER BY RANDOM() LIMIT 1")
	if err != nil {
		fmt.Printf("error preparing learn select: %v\n", err)
		return nil
	}

	return &LearnCommand{rtm, db, ins, del, sel}
}
