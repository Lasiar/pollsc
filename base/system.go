package base

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// Config main configure pollsc
type Config struct {
	VkToken string `json:"vk_token"`
	GroupID int    `json:"group_id"`
}

var (
	_config     *Config
	_onceConfig sync.Once
)

// GetConfig get object Config
func GetConfig() *Config {
	_onceConfig.Do(func() {
		_config = new(Config)
		_config.load()
	})
	return _config
}

func (c *Config) load() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	confFile, err := os.Open("/etc/pollsc/prod.Config.json")
	if err != nil {
		log.Fatal(err)
	}

	dc := json.NewDecoder(confFile)
	if err := dc.Decode(&c); err != nil {
		log.Fatal("Read Config file: ", err)
	}
}
