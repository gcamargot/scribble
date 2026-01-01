package handlers

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nahtao97/scribble/internal/models"
	"github.com/nahtao97/scribble/internal/services"
)

// LeaderboardHandler handles HTTP requests for leaderboard endpoints
type LeaderboardHandler struct {
	leaderboardService *services.LeaderboardService
}

// NewLeaderboardHandler creates a new leaderboard handler instance
func NewLeaderboardHandler(leaderboardService *services.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
	}
}

// ComputeLeaderboards handles POST /internal/leaderboards/compute
// Admin-only endpoint to trigger leaderboard computation
func (h *LeaderboardHandler) ComputeLeaderboards(c *gin.Context) {
	// Verify admin secret
	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "admin secret not configured",
		})
		return
	}

	providedSecret := c.GetHeader("X-Admin-Secret")
	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(providedSecret), []byte(adminSecret)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid admin credentials",
		})
		return
	}

	// Check if specific metric requested
	metricParam := c.Query("metric")

	if metricParam != "" {
		// Compute specific metric
		metricType := models.MetricType(metricParam)

		// Validate metric type
		valid := false
		for _, mt := range models.AllMetricTypes() {
			if mt == metricType {
				valid = true
				break
			}
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "invalid metric type",
				"valid_metrics": models.AllMetricTypes(),
			})
			return
		}

		result, err := h.leaderboardService.ComputeLeaderboard(metricType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"results": []models.ComputeResult{*result},
		})
		return
	}

	// Compute all metrics
	results, err := h.leaderboardService.ComputeAllLeaderboards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": results,
	})
}

// GetLeaderboard handles GET /internal/leaderboards/:metric
// Returns paginated leaderboard for a specific metric
// Query params: page, page_size
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	metricParam := c.Param("metric")
	metricType := models.MetricType(metricParam)

	// Validate metric type
	valid := false
	for _, mt := range models.AllMetricTypes() {
		if mt == metricType {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "invalid metric type",
			"valid_metrics": models.AllMetricTypes(),
		})
		return
	}

	// Parse pagination
	page := 1
	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if sizeParam := c.Query("page_size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	result, err := h.leaderboardService.GetLeaderboard(metricType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUserRanks handles GET /internal/leaderboards/user/:user_id
// Returns all ranks for a specific user
func (h *LeaderboardHandler) GetUserRanks(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	ranks, err := h.leaderboardService.GetUserAllRanks(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"ranks":   ranks,
	})
}

// GetAvailableMetrics handles GET /internal/leaderboards/metrics
// Returns list of available metric types
func (h *LeaderboardHandler) GetAvailableMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"metrics": models.AllMetricTypes(),
	})
}
