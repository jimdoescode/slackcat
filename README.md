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

### Building

```bash
$ go build
```

### Running

```bash
$ slackcat <SLACKBOT_TOKEN> <YOUR_SLACK_USERID>
```

### Dependencies
- [golang](https://golang.org/)
- [sqlite](https://www.sqlite.org/)
- [slack api](https://godoc.org/github.com/nlopes/slack)
- [go-git](https://godoc.org/gopkg.in/src-d/go-git.v4)


Commands
--------

- **Learn** `Syntax: ?(un)learn <target> <value>` 

  Is a way of associating text to a particular target. Then randomly recalling the text whenever the target is queried.
- **Plus** `Syntax: ?++|-- <target>` 

  Is a way of giving arbitrary internet points to a target.
- **Giphy** `Syntax: ?giphy <search query>`

  Does a standard giphy search.
- **Update** `Syntax: ?update`

  Will pull down the latest changes to slack cat and rebuild the binary then shut slack cat down so it can be restarted.
- **Halt** `Syntax: ?halt`

  Will stop the currently executing slack cat process. Depending upon set up it may be restarted automatically.
*Multiple commands cannot currently be combined for safety reasons.*
