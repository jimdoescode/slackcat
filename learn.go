package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

type LearnCommand struct {
	db  *sql.DB
	ins *sql.Stmt
	del *sql.Stmt
	sel *sql.Stmt
}

func (c *LearnCommand) Execute(msg *Message) (*Message, error) {

	txt := strings.SplitN(msg.Text, " ", 3)

	if len(txt[0]) < 2 {
		return nil, nil
	}

	token := strings.ToLower(txt[0][1:])

	if token != "learn" && token != "unlearn" {
		var val string
		err := c.sel.QueryRow(token).Scan(&val)
		if err != nil {
			fmt.Printf("error searching db: %v\n", err)
			return nil, nil
		}

		msg.Text = val
		return msg, err
	}

	if len(txt) < 3 {
		msg.Text = "Syntax: ?(un)learn <target> <value>"
		return msg, nil
	}

	if token == "learn" {
		_, err := c.ins.Exec(strings.ToLower(txt[1]), txt[2])
		msg.Text = fmt.Sprintf("OK, learned %s", txt[1])

		fmt.Printf("%s %s", txt[1], txt[2])

		return msg, err
	}

	if token == "unlearn" {
		_, err := c.del.Exec(strings.ToLower(txt[1]), txt[2])
		msg.Text = fmt.Sprintf("Unlearned %s", txt[1])
		return msg, err
	}

	return nil, nil
}

func (c *LearnCommand) Close() {
	c.sel.Close()
	c.del.Close()
	c.ins.Close()
	c.db.Close()
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
