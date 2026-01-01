package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nahtao97/scribble/internal/services"
)

// UserHandler handles HTTP requests for user-related endpoints
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler instance
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserMetrics handles GET /internal/users/:user_id/metrics
// Returns aggregate metrics for a specific user
func (h *UserHandler) GetUserMetrics(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	metrics, err := h.userService.GetUserMetrics(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetUserMetricsByUsername handles GET /internal/users/username/:username/metrics
// Returns aggregate metrics for a user by username
func (h *UserHandler) GetUserMetricsByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username is required",
		})
		return
	}

	metrics, err := h.userService.GetUserMetricsByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetTopUsers handles GET /internal/users/top
// Returns top users by problems solved or streak
// Query params: by (problems|streak), limit (default 10, max 100)
func (h *UserHandler) GetTopUsers(c *gin.Context) {
	by := c.DefaultQuery("by", "problems")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var users interface{}

	switch by {
	case "problems":
		users, err = h.userService.GetTopUsersByProblems(limit)
	case "streak":
		users, err = h.userService.GetTopUsersByStreak(limit)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "invalid 'by' parameter",
			"valid_options": []string{"problems", "streak"},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"metric": by,
		"limit":  limit,
	})
}

// GetUserLanguageStats handles GET /internal/users/:user_id/languages
// Returns language usage statistics for a user
func (h *UserHandler) GetUserLanguageStats(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	stats, err := h.userService.GetLanguageStats(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":   userID,
		"languages": stats,
	})
}
