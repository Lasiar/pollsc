package main

import (
	"fmt"
	"log"
	"os"
	"terminal-vk/VK"
	"terminal-vk/base"
	"terminal-vk/client"
)

func main() {

	bot := VK.New(base.GetConfig().VkToken, "5.92")

	bot.Debug = true
	logger := log.New(os.Stderr, "VK ", log.LstdFlags)

	bot.SetLogger(logger)

	srv, err := bot.GetLongPoolServer(base.GetConfig().GroupID)
	if err != nil {
		log.Println(err)
	}

	updates := srv.Listen()

	message := client.Init()

	for {
		select {
		case m := <-message:
			if err := bot.MessagesSend(m.Text, m.ID); err != nil {
				log.Println(err)
			}

		case update := <-updates:
			err := client.Processed(update.Object.Text, update.Object.FromID)
			if err := bot.MessagesSend("Added", update.Object.FromID); err != nil {
				log.Println(err)
				continue
			}

			if err := bot.MessagesSend(fmt.Sprint(err), update.Object.FromID); err != nil {
				log.Println(err)
				continue
			}

		}
	}
}

