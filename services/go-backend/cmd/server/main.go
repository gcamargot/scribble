package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	leaderboardService := services.NewLeaderboardService(database.GetConnection())
	antiCheatService := services.NewAntiCheatService(database.GetConnection())
	submissionService := services.NewSubmissionService(database.GetConnection())

	// Initialize handlers
	problemHandler := handlers.NewProblemHandler(problemService)
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)
	antiCheatHandler := handlers.NewAntiCheatHandler(antiCheatService)
	submissionHandler := handlers.NewSubmissionHandler(submissionService)

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

		// Leaderboard endpoints
		leaderboards := internal.Group("/leaderboards")
		{
			// GET /internal/leaderboards/metrics - List available metric types
			leaderboards.GET("/metrics", leaderboardHandler.GetAvailableMetrics)

			// POST /internal/leaderboards/compute - Compute rankings (admin only)
			// Query param: metric to compute specific metric only
			leaderboards.POST("/compute", leaderboardHandler.ComputeLeaderboards)

			// GET /internal/leaderboards/user/:user_id - Get user's ranks
			leaderboards.GET("/user/:user_id", leaderboardHandler.GetUserRanks)

			// GET /internal/leaderboards/:metric - Get paginated leaderboard
			// Query params: page, page_size
			leaderboards.GET("/:metric", leaderboardHandler.GetLeaderboard)
		}

		// Submission endpoints
		submissions := internal.Group("/submissions")
		{
			// GET /internal/submissions/:id - Get submission by ID
			// Query param: include_code=true to include source code
			submissions.GET("/:id", submissionHandler.GetSubmissionByID)

			// GET /internal/submissions/user/:user_id - Get user submission history
			// Query params: page, page_size, problem_id, status, language
			submissions.GET("/user/:user_id", submissionHandler.GetUserSubmissionHistory)

			// GET /internal/submissions/user/:user_id/stats - Get user submission stats
			submissions.GET("/user/:user_id/stats", submissionHandler.GetUserSubmissionStats)
		}

		// Anti-cheat endpoints
		anticheat := internal.Group("/anticheat")
		{
			// POST /internal/anticheat/check - Check if submission should be allowed
			anticheat.POST("/check", antiCheatHandler.CheckSubmission)

			// POST /internal/anticheat/flag - Flag a submission
			anticheat.POST("/flag", antiCheatHandler.FlagSubmission)
		}

		// Admin endpoints (require X-Admin-Secret header)
		admin := internal.Group("/admin")
		{
			// GET /internal/admin/flags/pending - Get pending flags
			admin.GET("/flags/pending", antiCheatHandler.GetPendingFlags)

			// GET /internal/admin/flags/stats - Get flag statistics
			admin.GET("/flags/stats", antiCheatHandler.GetFlagStats)

			// GET /internal/admin/flags/user/:user_id - Get flags for user
			admin.GET("/flags/user/:user_id", antiCheatHandler.GetUserFlags)

			// POST /internal/admin/flags/:flag_id/review - Review a flag
			admin.POST("/flags/:flag_id/review", antiCheatHandler.ReviewFlag)

			// POST /internal/admin/cleanup/rate-limits - Cleanup stale entries
			admin.POST("/cleanup/rate-limits", antiCheatHandler.CleanupRateLimits)
		}

		// TODO: Add streak endpoints
	}

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

	// Check if mTLS is enabled
	tlsCert := os.Getenv("TLS_CERT_PATH")
	tlsKey := os.Getenv("TLS_KEY_PATH")
	tlsCA := os.Getenv("TLS_CA_PATH")
	mtlsEnabled := tlsCert != "" && tlsKey != "" && tlsCA != ""

	// Start server in a goroutine so it doesn't block shutdown handling
	go func() {
		if mtlsEnabled {
			// Configure mTLS
			caCert, err := os.ReadFile(tlsCA)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read CA cert: %v\n", err)
				os.Exit(1)
			}

			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				fmt.Fprintf(os.Stderr, "Failed to parse CA cert\n")
				os.Exit(1)
			}

			srv.TLSConfig = &tls.Config{
				ClientCAs:  caCertPool,
				ClientAuth: tls.RequireAndVerifyClientCert,
				MinVersion: tls.VersionTLS12,
			}

			fmt.Printf("Starting Go backend server with mTLS on port %s (env: %s)\n", port, os.Getenv("GO_ENV"))
			if err := srv.ListenAndServeTLS(tlsCert, tlsKey); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Starting Go backend server on port %s (env: %s)\n", port, os.Getenv("GO_ENV"))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				os.Exit(1)
			}
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
