package handlers

import (
	"github.com/gin-gonic/gin"
)

// Health represents the health check handler
// This is a placeholder for the health endpoint handler
// Currently implemented in the server package for simplicity
type Health struct {
	// Can be expanded with dependencies like database, cache, etc.
}

// NewHealth creates a new health handler
func NewHealth() *Health {
	return &Health{}
}

// Check returns the health status of the service and its dependencies
// This can be expanded to check database, Kubernetes, executor connections, etc.
func (h *Health) Check(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "scribble-go-backend",
	})
}
