package base

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type config struct {
	VkToken string `json:"vk_token"`
	GroupID int    `json:"group_id"`
}
var (
	_config     *config
	_onceConfig sync.Once
)

//GetConfig get object config
func GetConfig() *config {
	_onceConfig.Do(func() {
		_config = new(config)
		_config.load()
	})
	return _config
}

func (c *config) load() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	confFile, err := os.Open("/etc/pollsc/prod.config.json")
	if err != nil {
		log.Fatal(err)
	}

	dc := json.NewDecoder(confFile)
	if err := dc.Decode(&c); err != nil {
		log.Fatal("Read config file: ", err)
	}
}
