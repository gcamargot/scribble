package server

import (
	"fmt"
	"os"
)

// Config holds the server configuration
type Config struct {
	Port     string
	Env      string
	LogLevel string
}

// NewConfig creates a new server configuration from environment variables
func NewConfig() *Config {
	return &Config{
		Port:     getEnvOrDefault("GO_PORT", "8080"),
		Env:      getEnvOrDefault("GO_ENV", "development"),
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),
	}
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	return fmt.Sprintf(
		"Server Config - Port: %s, Env: %s, LogLevel: %s",
		c.Port, c.Env, c.LogLevel,
	)
}
