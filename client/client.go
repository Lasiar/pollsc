/*
Wrapper over server
*/

package client

import (
	"net/url"
	"strings"
	"terminal-vk/server"
	"time"
)

var (
	add chan server.Client
)

type Message struct {
	Text string
	ID   int
}

func Init() chan Message {
	var out chan server.Message

	outMessage := make(chan Message)
	add, _, out = server.Start()
	go middleChangeState(out, outMessage)
	return outMessage
}

func middleChangeState(out chan server.Message, outMessage chan Message) {
	for o := range out {
		outMessage <- Message{o.Text, o.Client.ClientID}
	}
}

func Processed(message string, id int) error {

	firstSpace := strings.Index(message, " ")
	firstWord := message[:firstSpace]

	if len(message)+2 == firstSpace {
		return nil
	}

	arguments := message[firstSpace:]

	if firstWord == "listen" {
		return addToListen(arguments, id)
	}

	return nil
}

func addToListen(args string, id int) error {
	var err error

	urls := strings.Split(strings.TrimSpace(args), ",")

	validUrls := make([]*url.URL, len(urls), len(urls))

	for i, current := range urls {
		validUrls[i], err = url.Parse(current)

		if validUrls[i].Scheme == "" {
			return err
		}
		add <- server.Client{ClientID: id, Url: *validUrls[i], Timer: time.Duration(10 * time.Second)}
	}

	return nil
}
