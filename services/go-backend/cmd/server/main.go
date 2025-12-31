package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file if it exists
	_ = godotenv.Load()
}

func main() {
	// Set Gin mode
	if os.Getenv("GO_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()

	// Health check endpoint
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

	// Start server
	port := os.Getenv("GO_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting Go backend on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
