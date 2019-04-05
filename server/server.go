package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// State status polling service
type State struct {
	URL string
}

// Client object for service
type Client struct {
	IsGood   bool
	ClientID int
	URL      url.URL
	Timer    time.Duration
}

// Message message for user
type Message struct {
	Client     Client
	HTTPStatus string
	Text       string
}

type storage map[int]struct {
	URL url.URL
	ch  chan bool
}

var storages storage

func init() {
	storages = make(storage)
}

// Start start work server
func Start() (addClient chan Client, deleteClient chan int, message chan Message) {
	message = make(chan Message)
	addClient, deleteClient = make(chan Client), make(chan int)

	go worker(addClient, deleteClient, message)

	return addClient, deleteClient, message
}

func worker(addClients chan Client, deleteChan chan int, out chan Message) {
	for {
		select {

		case client := <-addClients:
			fmt.Println(client)
			deleteChan := make(chan bool)

			storages[client.ClientID] = struct {
				URL url.URL
				ch  chan bool
			}{URL: client.URL, ch: deleteChan}

			go client.checker(out, deleteChan)

		case id := <-deleteChan:
			storages[id].ch <- true
			delete(storages, id)
		}
	}
}

// GetInfo get subscriber service for user
func GetInfo(id int) string {
	st := storages[id]
	return st.URL.String()
}

func (c Client) checker(out chan Message, deleteChan chan bool) {
	timer := time.NewTimer(c.Timer)
	httpClient := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	for {
		select {
		case <-timer.C:
			resp, err := httpClient.Get(c.URL.String())
			log.Println("http get ", c.URL)

			timer.Reset(c.Timer)

			if err != nil {

				if c.IsGood {
					c.IsGood = false
					out <- Message{Client: c, Text: fmt.Sprint(err)}
				}
				continue
			}

			// All ok`
			if resp.StatusCode < 400 && !c.IsGood {
				c.IsGood = true
				out <- Message{Client: c, Text: "Появилась в сети", HTTPStatus: resp.Status}
				continue
			}

			// Error
			if resp.StatusCode > 400 && c.IsGood {
				c.IsGood = false
				out <- Message{Client: c, Text: "Исчез из сети", HTTPStatus: resp.Status}
				continue
			}
		case <-deleteChan:
			fmt.Println("remove: ", c.ClientID)
			return
		}
	}
}
