package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
	"strings"
)

type LearnCommand struct {
	db  *sql.DB
	ins *sql.Stmt
	del *sql.Stmt
	sel *sql.Stmt
}

func (c *LearnCommand) Execute(msg *SlackMessage) (*SlackMessage, error) {
	txt := strings.SplitN(msg.Text, " ", 3)
	token := strings.ToLower(txt[0][1:])

	if token != "learn" && token != "unlearn" {
		var val string
		err := c.sel.QueryRow(token).Scan(&val)
		if err != nil {
			fmt.Printf("error searching db: %v\n", err)
			return nil, nil
		}

		msg.Text = c.parseText(val)
		return msg, err
	}

	if len(txt) < 3 {
		msg.Text = c.GetSyntax()
		return msg, nil
	}

	if token == "learn" {
		target := strings.ToLower(txt[1])

		if target == token {
			msg.Text = "We must go deeper!"
			return msg, nil
		}

		_, err := c.ins.Exec(target, txt[2])
		msg.Text = fmt.Sprintf("OK, learned %s", txt[1])

		return msg, err
	}

	if token == "unlearn" {
		target := strings.ToLower(txt[1])

		if target == token {
			msg.Text = "Don't incept me!"
			return msg, nil
		}

		_, err := c.del.Exec(target, txt[2])
		msg.Text = fmt.Sprintf("Unlearned %s", txt[1])

		return msg, err
	}

	return nil, nil
}

func (c *LearnCommand) GetSyntax() string {
	return "Syntax: ?(un)learn <target> <value>"
}

func (c *LearnCommand) Close() {
	c.sel.Close()
	c.del.Close()
	c.ins.Close()
	c.db.Close()
}

func (c *LearnCommand) parseText(txt string) string {
	re := regexp.MustCompile("\\?([^\\s]+)")
	vars := re.FindAllStringSubmatch(txt, -1)

	for _, v := range vars {
		var val string
		err := c.sel.QueryRow(v[1]).Scan(&val)
		if err != nil {
			val = v[0]
		}

		txt = strings.Replace(txt, v[0], val, 1)
	}

	return txt
}

func NewLearnCommand() *LearnCommand {
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

	return &LearnCommand{db, ins, del, sel}
}
