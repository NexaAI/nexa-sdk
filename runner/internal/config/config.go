package config

import (
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Host string
}

var config *Config
var once sync.Once

func Get() *Config {
	once.Do(get)
	return config
}

func init() {
	viper.SetDefault("host", "127.0.0.1:18181")

}

func get() {
	config = &Config{}

	viper.SetEnvPrefix("nexa")
	viper.AutomaticEnv()
	viper.Unmarshal(config)
}
