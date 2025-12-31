package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/nahtao97/scribble/internal/server"
)

func init() {
	// Load .env file if it exists (for local development)
	// Ignore error if .env doesn't exist (it's optional)
	_ = godotenv.Load()
}

func main() {
	// Load server configuration from environment variables
	config := server.NewConfig()

	// Create server instance
	srv := server.NewServer(config)

	// Start the server
	// This will block until the server is stopped or encounters a fatal error
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}
