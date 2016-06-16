package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	//"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type PlusDenominationCommand struct {
	db  *sql.DB
	ins *sql.Stmt
	del *sql.Stmt
	sel *sql.Stmt
}

func (c *PlusDenominationCommand) Execute(msg *SlackMessage) (*SlackMessage, error) {
	txt := strings.SplitN(msg.Text, " ", 3)
	token := strings.ToLower(txt[0][1:])

	if token != "++d" && token != "--d" {
		return nil, nil
	}

	if len(txt) == 1 {
		var err error
		msg.Text, err = c.getDenominationsDisplay()
		return msg, err
	} else if len(txt) < 3 {
		msg.Text = c.GetSyntax()
		return msg, nil
	}

	idx, err := strconv.Atoi(txt[1])
	if err != nil {
		msg.Text = c.GetSyntax()
		return msg, err
	}

	if idx == 0 {
		msg.Text = fmt.Sprintf("0 ain't no denomination!")
		return msg, nil
	}

	c.del.Exec(idx)

	if token == "++d" {
		_, err := c.ins.Exec(idx, txt[2])
		if err != nil {
			msg.Text = c.GetSyntax()
			return msg, err
		}
		msg.Text = fmt.Sprintf("OK, added plus denomination %s", txt[2])
	} else {
		msg.Text = fmt.Sprintf("OK, removed plus denomination %s", txt[2])
	}

	return msg, nil
}

func (c *PlusDenominationCommand) getDenominationsDisplay() (string, error) {
	rows, err := c.sel.Query()
	defer rows.Close()
	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("Here's the current plus exchange rate\n```")
	w := tabwriter.NewWriter(buf, 7, 0, 1, ' ', 0)
	for rows.Next() {
		var val int
		var name string
		err = rows.Scan(&val, &name)
		if err != nil {
			return "", err
		}

		fmt.Fprintf(w, "%d:\t%s\n", val, name)
	}
	fmt.Fprint(w, "```")
	w.Flush()
	return buf.String(), nil
}

func (c *PlusDenominationCommand) GetSyntax() string {
	return "syntax: ?(++|--)d <plus count> <name>"
}

func (c *PlusDenominationCommand) Close() {
	c.sel.Close()
	c.ins.Close()
	c.del.Close()
	c.db.Close()
}

func NewPlusDenominationCommand() *PlusDenominationCommand {
	db, err := sql.Open("sqlite3", "./slackcat.db")
	if err != nil {
		fmt.Printf("error creating plus command: %v\n", err)
		return nil
	}

	db.Exec("CREATE TABLE plus_denominations (value INTEGER PRIMARY KEY NOT NULL, name TEXT)")

	ins, err := db.Prepare("INSERT INTO plus_denominations(value, name) VALUES(?,?)")
	if err != nil {
		fmt.Printf("error preparing plus_denominations insert: %v\n", err)
		return nil
	}

	del, err := db.Prepare("DELETE from plus_denominations WHERE value=?")
	if err != nil {
		fmt.Printf("error preparing plus_denominations delete: %v\n", err)
		return nil
	}

	sel, err := db.Prepare("SELECT * FROM plus_denominations ORDER BY value ASC")
	if err != nil {
		fmt.Printf("error preparing plus_denominations select: %v\n", err)
		return nil
	}

	return &PlusDenominationCommand{db, ins, del, sel}
}
