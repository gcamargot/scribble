package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()
}

func main() {
	// Set Gin mode based on environment
	if os.Getenv("GO_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()

	// Health check endpoint for Kubernetes probes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "scribble-go-backend",
		})
	})

	// TODO: Add more routes here
	// - Problem retrieval endpoints
	// - Code execution endpoints
	// - Leaderboard endpoints
	// - Streak endpoints

	// Configure HTTP server
	port := os.Getenv("GO_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
		// Timeouts prevent slow clients from holding connections indefinitely
		ReadTimeout:  15 * time.Second, // Time to read request headers + body
		WriteTimeout: 15 * time.Second, // Time to write response
		IdleTimeout:  60 * time.Second, // Time to keep connection alive between requests
	}

	// Channel to listen for interrupt/termination signals
	quit := make(chan os.Signal, 1)
	// Catch SIGINT (Ctrl+C) and SIGTERM (Kubernetes pod termination)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine so it doesn't block shutdown handling
	go func() {
		fmt.Printf("Starting Go backend server on port %s (env: %s)\n", port, os.Getenv("GO_ENV"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Block until we receive a signal
	<-quit
	fmt.Println("\nReceived shutdown signal, initiating graceful shutdown...")

	// Create context with timeout for graceful shutdown
	// Give in-flight requests 15 seconds to complete before forcing shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	// This will:
	// 1. Stop accepting new requests
	// 2. Wait for existing requests to complete (up to 15s)
	// 3. Close all connections
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error during graceful shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server shutdown complete. All connections closed gracefully.")
}
