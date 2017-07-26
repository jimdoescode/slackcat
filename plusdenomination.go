package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nlopes/slack"
	"strconv"
	"strings"
	"text/tabwriter"
)

type PlusDenominationCommand struct {
	rtm *slack.RTM
	db  *sql.DB
	ins *sql.Stmt
	del *sql.Stmt
	sel *sql.Stmt
}

func (c *PlusDenominationCommand) Matches(msg *slack.Msg) bool {
	return strings.HasPrefix(msg.Text, "?++d") ||
		strings.HasPrefix(msg.Text, "?--d")
}

func (c *PlusDenominationCommand) Execute(msg *slack.Msg) (*slack.OutgoingMessage, error) {
	txt := strings.SplitN(msg.Text, " ", 3)
	token := strings.ToLower(txt[0][1:])

	if len(txt) == 1 {
		disp, err := c.getDenominationsDisplay()
		out := c.rtm.NewOutgoingMessage(disp, msg.Channel)
		return out, err
	} else if len(txt) < 3 {
		disp := c.GetSyntax()
		out := c.rtm.NewOutgoingMessage(disp, msg.Channel)
		return out, nil
	}

	idx, err := strconv.Atoi(txt[1])
	if err != nil {
		disp := c.GetSyntax()
		out := c.rtm.NewOutgoingMessage(disp, msg.Channel)
		return out, err
	}

	if idx == 0 {
		out := c.rtm.NewOutgoingMessage("0 ain't no denomination!", msg.Channel)
		return out, nil
	}

	c.del.Exec(idx)

	if token == "++d" {
		_, err := c.ins.Exec(idx, txt[2])
		if err != nil {
			disp := c.GetSyntax()
			out := c.rtm.NewOutgoingMessage(disp, msg.Channel)
			return out, err
		}
		out := c.rtm.NewOutgoingMessage(fmt.Sprintf("OK, added plus denomination %s", txt[2]), msg.Channel)
		return out, nil
	}

	out := c.rtm.NewOutgoingMessage(fmt.Sprintf("OK, removed plus denomination %s", txt[2]), msg.Channel)
	return out, nil
}

func (c *PlusDenominationCommand) getDenominationsDisplay() (string, error) {
	rows, err := c.sel.Query()
	if err != nil {
		rows.Close()
		return "", err
	}

	buf := bytes.NewBufferString("Here's the current plus exchange rate\n```")
	w := tabwriter.NewWriter(buf, 7, 0, 1, ' ', 0)
	for rows.Next() {
		var val int
		var name string
		err = rows.Scan(&val, &name)
		if err != nil {
			rows.Close()
			return "", err
		}

		fmt.Fprintf(w, "%d:\t%s\n", val, name)
	}
	fmt.Fprint(w, "```")
	w.Flush()
	rows.Close()
	return buf.String(), nil
}

func (c *PlusDenominationCommand) GetSyntax() string {
	return "?(++|--)d <plus count> <name>"
}

func (c *PlusDenominationCommand) Close() {
	c.sel.Close()
	c.ins.Close()
	c.del.Close()
	c.db.Close()
}

func NewPlusDenominationCommand(rtm *slack.RTM) *PlusDenominationCommand {
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

	return &PlusDenominationCommand{rtm, db, ins, del, sel}
}
