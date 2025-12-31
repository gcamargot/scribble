package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nahtao97/scribble/internal/db"
	"github.com/nahtao97/scribble/internal/handlers"
	"github.com/nahtao97/scribble/internal/services"
)

func init() {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()
}

func main() {
	// Initialize database connection
	database, err := db.NewDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Verify database connection
	if err := database.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Database ping failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database connection established")

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

	// Initialize services
	problemService := services.NewProblemService(database.GetConnection())

	// Initialize handlers
	problemHandler := handlers.NewProblemHandler(problemService)

	// Register API routes under /internal prefix
	// These endpoints are called by the Node.js proxy (internal service-to-service)
	internal := router.Group("/internal")
	{
		// Problem endpoints
		problems := internal.Group("/problems")
		{
			// GET /internal/problems/daily/:date - Get daily challenge
			// :date can be "today" or YYYY-MM-DD format
			problems.GET("/daily/:date", problemHandler.GetDailyChallengeByDate)

			// GET /internal/problems/:id - Get specific problem by ID
			problems.GET("/:id", problemHandler.GetProblemByID)

			// GET /internal/problems/:id/test-cases - Get test cases for a problem
			// Query param: all=true to include hidden tests (for code executor)
			problems.GET("/:id/test-cases", problemHandler.GetTestCasesByProblemID)
		}

		// TODO: Add submission endpoints
		// TODO: Add leaderboard endpoints
		// TODO: Add streak endpoints
	}

	// Start server
	port := os.Getenv("GO_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting Go backend server on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
