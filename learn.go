package main

import (
	"database/sql"
	"fmt"
	"github.com/nlopes/slack"
	"regexp"
	"strings"
)

type LearnCommand struct {
	rtm *slack.RTM
	exp *regexp.Regexp
	ins *sql.Stmt
	del *sql.Stmt
	sel *sql.Stmt
}

func (c *LearnCommand) Matches(msg *slack.Msg) (bool, bool) {
	if c.exp.MatchString(msg.Text) {
		return true, false
	}

	txt := strings.SplitN(msg.Text, " ", 2)
	if len(txt) < 1 || len(txt[0]) < 1 || txt[0][0] != '?' {
		return false, false
	}

	token := c.parseTarget(
		strings.ToLower(txt[0][1:]),
	)

	var val string
	err := c.sel.QueryRow(token).Scan(&val)

	return err == nil, false
}

func (c *LearnCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	if c.exp.MatchString(msg.Text) {
		vars := c.exp.FindStringSubmatch(msg.Text)
		dbCmd := c.ins
		target := c.parseTarget(vars[2])

		out := c.rtm.NewOutgoingMessage(fmt.Sprintf("OK, learned %s", target), msg.Channel)
		if vars[1] == "unlearn" {
			dbCmd = c.del
			out.Text = fmt.Sprintf("Unlearned %s", target)
		}

		_, err := dbCmd.Exec(target, vars[3])
		return out, err
	}

	txt := strings.SplitN(msg.Text, " ", 2)
	token := c.parseTarget(
		strings.ToLower(txt[0][1:]),
	)

	var val string
	err := c.sel.QueryRow(token).Scan(&val)
	if err != nil {
		fmt.Printf("error searching db: %v\n", err)
		return nil, nil
	}

	out := c.rtm.NewOutgoingMessage(c.parseText(val), msg.Channel)
	return out, nil
}

func (c *LearnCommand) GetSyntax() string {
	return "?(un)learn <target> <value>"
}

func (c *LearnCommand) GetDescription() string {
	return "Make slack cat associate two things. These associations can be randomly displayed by typing `?<target>`"
}

func (c *LearnCommand) Close() {
	c.sel.Close()
	c.del.Close()
	c.ins.Close()
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

func NewLearnCommand(rtm *slack.RTM, db *sql.DB) *LearnCommand {
	exp := regexp.MustCompile(`^(?i)\?(learn|unlearn) ([\w@<>\|#]+) (.+?)$`)

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

	return &LearnCommand{rtm, exp, ins, del, sel}
}
