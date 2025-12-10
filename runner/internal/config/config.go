package config

import (
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	// Server settings
	Host      string // Server host and port (default: "127.0.0.1:18181")
	Origins   string // Allowed CORS origins (default: "*")
	KeepAlive int64  // Connection keep-alive timeout in seconds (default: 300)
	// HTTPS / TLS settings
	EnableHTTPS bool   // Whether to serve over HTTPS (default: false)
	CertFile    string // TLS certificate file path
	KeyFile     string // TLS private key file path

	// Env only params
	HFToken string
	Log     string
}

var config *Config
var once sync.Once

// Get returns the singleton configuration instance.
// Uses sync.Once to ensure configuration is loaded only once.
func Get() *Config {
	once.Do(get)
	return config
}

// NOTE: Avoid calling Get before subcommand initialization to prevent premature config initialization
func GetLog() string {
	get()
	return config.Log
}

// init sets up default configuration values using Viper.
// These defaults are used if no environment variables are provided.
func init() {
	// ENV only param need to set default here
	viper.SetDefault("hftoken", "") // Default empty token
	viper.SetDefault("log", "none") // Default log level
}

// get initializes the configuration by reading from environment variables.
// Environment variables should be prefixed with "NEXA_" (e.g., NEXA_HOST).
// This function is called only once via sync.Once for thread safety.
func get() {
	config = &Config{}

	// Set environment variable prefix to "NEXA_"
	viper.SetEnvPrefix("nexa")
	// Automatically read environment variables
	viper.AutomaticEnv()
	// Unmarshal configuration into the Config struct
	viper.Unmarshal(config)
}
