package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"terminal-vk/VK"
	"terminal-vk/base"
	"terminal-vk/server"
	"time"
)

func main() {

	add, _, out := server.Test()

	bot := VK.New(base.GetConfig().VkToken, "5.92")

	bot.Debug = true
	logger := log.New(os.Stderr, "VK ", log.LstdFlags)

	bot.SetLogger(logger)

	srv, err := bot.GetLongPoolServer(base.GetConfig().GroupID)
	if err != nil {
		log.Println(err)
	}

	updates := srv.Listen()

	for {
		select {
		case o := <-out:
			if err := bot.MessagesSend(fmt.Sprintf("%v, %v, %v", o.Client.Url, o.Text, o.HttpStatus), o.Client.ClientID); err != nil {
				log.Println(err)
			}
		case update := <-updates:
			if strings.Contains(update.Object.Text, "listen") {
				parsed := strings.Split(update.Object.Text, " ")
				if len(parsed) != 2 {
					if err := bot.MessagesSend("Error: please input: command: arg", update.Object.FromID); err != nil {
						log.Println(err)
					}
					continue
				}

				url, err := url.Parse(parsed[1])
				if err != nil || url.Scheme == "" {
					if err := bot.MessagesSend("Error: please input correct url "+fmt.Sprint(err), update.Object.FromID); err != nil {
						log.Println(err)
					}
					continue
				}

				add <- server.Client{Url: *url, ClientID: update.Object.FromID, Timer: time.Duration(10 * time.Second)}

				if err := bot.MessagesSend("Added", update.Object.FromID); err != nil {
					log.Println(err)
				}

			}
		}
	}
}
