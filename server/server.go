package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type State struct {
	timer   *time.Timer
	current int
	id      int
	Url     string
}

type Client struct {
	IsGood   bool
	event    int
	ClientID int
	Url      url.URL
	Timer    time.Duration
}

type Message struct {
	Client     Client
	HttpStatus string
	Text       string
}

func Test() (chan Client, chan int, chan Message) {
	out := make(chan Message)
	AddClient, DeleteClient := make(chan Client), make(chan int)

	go Worker(AddClient, DeleteClient, out)

	return AddClient, DeleteClient, out
}

func Worker(AddClients chan Client, DeleteChan chan int, out chan Message) {
	c := make(map[int]chan bool)
	for {
		select {
		case client := <-AddClients:
			fmt.Println(client)
			deleteChan := make(chan bool)
			c[client.ClientID] = deleteChan
			go client.Cheker(out, deleteChan)
		case id := <-DeleteChan:
			c[id] <- true
			delete(c, id)
		}
	}
}

func (c Client) Cheker(out chan Message, DeleteChan chan bool) {
	timer := time.NewTimer(c.Timer)
	httpClient := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	for {
		select {
		case <-timer.C:
			resp, err := httpClient.Get(c.Url.String())
			log.Println("http get ", c.Url)

			timer.Reset(c.Timer)

			if err != nil {

				if c.IsGood == true {
					c.IsGood = false
					out <- Message{Client: c, Text: fmt.Sprint(err)}
				}
				continue
			}

			// All ok`
			if resp.StatusCode < 400 && !c.IsGood {
				c.IsGood = true
				out <- Message{Client: c, Text: "Появилась в сети", HttpStatus: resp.Status}
				continue
			}

			// Error
			if resp.StatusCode > 400 && c.IsGood {
				c.IsGood = false
				out <- Message{Client: c, Text: "Исчез из сети", HttpStatus: resp.Status}
				continue
			}

		case <-DeleteChan:
			fmt.Println("remove: ", c.ClientID)
			return
		}
	}
}
