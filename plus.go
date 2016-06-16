package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"sort"
	"strings"
)

type PlusCommand struct {
	db       *sql.DB
	ins      *sql.Stmt
	upd      *sql.Stmt
	sel      *sql.Stmt
	selDenom *sql.Stmt
}

func (c *PlusCommand) Execute(msg *SlackMessage) (*SlackMessage, error) {
	txt := strings.SplitN(msg.Text, " ", 3)
	token := strings.ToLower(txt[0][1:])

	if token != "++" && token != "--" {
		return nil, nil
	}

	if len(txt) < 2 {
		msg.Text = c.GetSyntax()
		return msg, nil
	}

	target := strings.ToLower(txt[1])
	//Strip user or channel symbols
	if target[0] == '@' || target[0] == '#' {
		target = target[1:]
	}

	var val int
	err := c.sel.QueryRow(target).Scan(&val)
	if err != nil {
		fmt.Printf("error searching db: %v\n", err)
		c.ins.Exec(target, 0)
		val = 0
	}

	add := (token == "++")
	owner := (target == msg.User || target == msg.User[1:])

	if add {
		val += 1
		if owner {
			msg.Text = "You'll go blind that way."
			return msg, nil
		}
	} else {
		val -= 1
	}

	_, err = c.upd.Exec(val, target)
	if err != nil {
		fmt.Printf("error updating db: %v\n", err)
	}

	msg.Text = c.getMessage(add, target, msg.User, val)

	return msg, err
}

func (c *PlusCommand) getMessage(add bool, target string, user string, val int) string {
	buf := bytes.NewBufferString("")
	if add {
		buf.WriteString(fmt.Sprintf("%s gave a plus to %s, ", user, target))
	} else {
		buf.WriteString(fmt.Sprintf("%s took a plus from %s, ", user, target))
	}

	buf.WriteString(fmt.Sprintf("%s now has %s.", target, c.pluralize(val, "plus")))
	denom := c.denominationEquivalent(val)
	if denom != "" {
		buf.WriteString(fmt.Sprintf("\n\nThat's equivalent to %s", denom))
	}

	return buf.String()
}

func (c *PlusCommand) denominationEquivalent(val int) string {
	buf := bytes.NewBufferString("")
	if c.selDenom == nil {
		return ""
	}

	rows, err := c.selDenom.Query()
	defer rows.Close()

	denoms := make(map[int]string)
	for rows.Next() {
		var val int
		var name string
		err = rows.Scan(&val, &name)
		if err != nil {
			return ""
		}

		denoms[val] = name
	}

	/*denoms := map[int]string{
		-5:   "Krossover Kredit",
		1:    "Schrute Buck",
		5:    "Stanley Nickel",
		10:   "Pizza Slice",
		25:   "Beer",
		50:   "Glitter Jar",
		100:  "Rubber Band",
		150:  "Slap Bracelet",
		250:  "Leprechaun",
		500:  "Presidential Fist Bump",
		1000: "Unicorn",
	}*/

	// To store the keys in slice in sorted order
	var keys []int
	for k := range denoms {
		keys = append(keys, k)
	}

	if val < 0 {
		sort.Sort(sort.IntSlice(keys))
	} else {
		sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	}

	for _, denom := range keys {
		coins := 0
		if val > 0 && denom > 0 {
			for denom <= val {
				val -= denom
				coins += 1
			}
		} else if val < 0 && denom < 0 {
			for denom >= val {
				val -= denom
				coins += 1
			}
		}

		if coins > 0 {
			if buf.Len() == 0 {
				buf.WriteString(c.pluralize(coins, denoms[denom]))
				continue
			}

			if val == 0 {
				buf.WriteString(" and ")
			} else {
				buf.WriteString(", ")
			}

			buf.WriteString(c.pluralize(coins, denoms[denom]))
		}
	}

	//If we are making an approximation due to denomination gaps.
	if val != 0 && buf.Len() != 0 {
		buf.WriteString(" and a little extra")
	}

	return buf.String()
}

func (c *PlusCommand) pluralize(val int, txt string) string {
	if val == 1 {
		return fmt.Sprintf("%d %s", val, txt)
	}

	if txt[len(txt)-1:] == "s" {
		return fmt.Sprintf("%d %ses", val, txt)
	}
	return fmt.Sprintf("%d %ss", val, txt)
}

func (c *PlusCommand) GetSyntax() string {
	return "Syntax: ?++|-- <target>"
}

func (c *PlusCommand) Close() {
	c.selDenom.Close()
	c.sel.Close()
	c.upd.Close()
	c.ins.Close()
	c.db.Close()
}

func NewPlusCommand() *PlusCommand {
	db, err := sql.Open("sqlite3", "./slackcat.db")
	if err != nil {
		fmt.Printf("error creating plus command: %v\n", err)
		return nil
	}

	db.Exec("CREATE TABLE pluses (target TEXT PRIMARY KEY NOT NULL, count INTEGER)")

	ins, err := db.Prepare("INSERT INTO pluses(target, count) VALUES(?,?)")
	if err != nil {
		fmt.Printf("error preparing plus insert: %v\n", err)
		return nil
	}

	upd, err := db.Prepare("UPDATE pluses SET count=? WHERE target=?")
	if err != nil {
		fmt.Printf("error preparing plus update: %v\n", err)
		return nil
	}

	sel, err := db.Prepare("SELECT count FROM pluses WHERE target=?")
	if err != nil {
		fmt.Printf("error preparing plus select: %v\n", err)
		return nil
	}

	selDenom, err := db.Prepare("SELECT * FROM plus_denominations")

	return &PlusCommand{db, ins, upd, sel, selDenom}
}
