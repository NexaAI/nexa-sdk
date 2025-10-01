package config

import (
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Host      string // Server host and port (default: "127.0.0.1:18181")
	KeepAlive int64  // Connection keep-alive timeout in seconds (default: 300)
	HFToken   string

	Log string // Enable backend log

	// HTTPS / TLS settings
	EnableHTTPS bool   // Whether to serve over HTTPS (default: false)
	CertFile    string // TLS certificate file path
	KeyFile     string // TLS private key file path
	UseNgrok    bool   // Whether to use ngrok for public HTTPS tunnel (default: false)
}

var config *Config
var once sync.Once

// Get returns the singleton configuration instance.
// Uses sync.Once to ensure configuration is loaded only once.
func Get() *Config {
	once.Do(get)
	return config
}

func Refresh() {
	once = sync.Once{}
	once.Do(get)
}

// init sets up default configuration values using Viper.
// These defaults are used if no environment variables are provided.
func init() {
	viper.SetDefault("host", "127.0.0.1:18181") // Default server address
	viper.SetDefault("keepalive", 300)          // Default 5-minute timeout
	viper.SetDefault("hftoken", "")             // Default empty token

	viper.SetDefault("log", "none") // Default log level

	// HTTPS defaults - disabled unless explicitly enabled
	viper.SetDefault("enablehttps", false)
	viper.SetDefault("certfile", "cert.pem")
	viper.SetDefault("keyfile", "key.pem")
	viper.SetDefault("ngrok", false)
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

	// Map the ngrok flag properly
	config.UseNgrok = viper.GetBool("ngrok")
}
