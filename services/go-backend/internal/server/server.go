package server

import (
	"fmt"

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

// Start starts the HTTP server on the configured port
func (s *Server) Start() error {
	// Register all routes before starting the server
	s.RegisterRoutes()

	addr := fmt.Sprintf(":%s", s.config.Port)
	fmt.Printf("Starting Go backend server on %s (env: %s)\n", addr, s.config.Env)

	if err := s.router.Run(addr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// GetRouter returns the underlying Gin router for additional configuration if needed
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
