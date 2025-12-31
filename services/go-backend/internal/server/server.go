package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server instance
type Server struct {
	router *gin.Engine
	config *Config
}

// NewServer creates a new server instance with the given configuration
func NewServer(config *Config) *Server {
	// Set Gin mode based on environment
	if config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router with default middleware
	router := gin.Default()

	return &Server{
		router: router,
		config: config,
	}
}

// RegisterRoutes registers all API routes
func (s *Server) RegisterRoutes() {
	// Health check endpoint for Kubernetes readiness/liveness probes
	s.router.GET("/health", s.healthHandler)

	// API version endpoint
	s.router.GET("/api/version", s.versionHandler)

	// TODO: Register problem endpoints
	// TODO: Register submission endpoints
	// TODO: Register leaderboard endpoints
	// TODO: Register streak endpoints
}

// healthHandler returns the health status of the server
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "scribble-go-backend",
		"env":     s.config.Env,
	})
}

// versionHandler returns the API version
func (s *Server) versionHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"version": "0.1.0",
		"service": "scribble-go-backend",
	})
}

// Start starts the HTTP server with graceful shutdown support
// Implements graceful shutdown for Kubernetes compatibility:
// - Listens for SIGTERM (K8s pod termination) and SIGINT (Ctrl+C)
// - Drains existing connections before shutting down
// - 15-second shutdown timeout (fits within K8s 30s termination grace period)
func (s *Server) Start() error {
	// Register all routes before starting the server
	s.RegisterRoutes()

	addr := fmt.Sprintf(":%s", s.config.Port)

	// Create HTTP server with timeouts to prevent resource exhaustion
	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,
		// ReadTimeout prevents slow header attacks (slowloris)
		ReadTimeout: 15 * time.Second,
		// WriteTimeout prevents slow response writes from holding connections
		WriteTimeout: 15 * time.Second,
		// IdleTimeout cleans up idle keep-alive connections
		IdleTimeout: 60 * time.Second,
	}

	// Setup graceful shutdown signal handling
	// Buffered channel prevents signal loss if server isn't ready
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine so we can listen for shutdown signals
	go func() {
		fmt.Printf("Starting Go backend server on port %s (env: %s)\n", s.config.Port, s.config.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Block until we receive a shutdown signal
	<-quit
	fmt.Println("\nReceived shutdown signal, initiating graceful shutdown...")

	// Create context with 15-second timeout for graceful shutdown
	// This allows in-flight requests to complete before forceful termination
	// 15s ensures we finish before Kubernetes 30s terminationGracePeriodSeconds
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("error during graceful shutdown: %w", err)
	}

	fmt.Println("Server shutdown complete. All connections closed gracefully.")
	return nil
}

// GetRouter returns the underlying Gin router for additional configuration if needed
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
