package config

import (
	"os"
	"sync"
)

type Config struct {
	Host    string
	HFToken string
}

var config *Config
var once sync.Once

func Get() *Config {
	once.Do(config.get)
	return config
}

func (c *Config) get() {
	c = &Config{}

	c.HFToken, _ = os.LookupEnv("HF_TOKEN")
}
