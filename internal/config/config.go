package config

import (
	"os"
)

// Config represents runtime options for the PlanForge server and CLI.
type Config struct {
	Host             string
	Port             string
	OpenBrowser      bool
	AnonymizeDefault bool
}

// LoadConfig loads configuration with environment variable fallbacks.
func LoadConfig() *Config {
	host := os.Getenv("PLANFORGE_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("PLANFORGE_PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Host:             host,
		Port:             port,
		OpenBrowser:      false,
		AnonymizeDefault: false,
	}
}
