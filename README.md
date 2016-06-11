Slackcat
========

This is an alternative to IRCCat which is an IRC chat bot. As the name implies this bot runs on Slack.

Installation
------------

Slackcat is written in Go so all you should need to do is:
```bash
$ go get github.com/jimdoescode/slackcat
```
Which will install slackcat to where ever you have go configured to install packages.

####Building
```bash
$ go build
```

####Running
```bash
$ slackcat <SLACKBOT_TOKEN>
```

####Dependencies
- [golang](https://golang.org/)
- [sqlite](https://www.sqlite.org/)


Commands
--------

- **Learn** `Syntax: ?(un)learn <target> <value>` 

  Is a way of associating text to a particular target. Then randomly recalling the text whenever the target is queried.
- **Plus** `Syntax: ?++|-- <target>` 

  Is a way of giving arbitrary internet points to a target.

*Multiple commands cannot currently be combined for safety reasons.*
