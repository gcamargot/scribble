package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nahtao97/scribble/internal/services"
)

// StreakHandler handles streak-related HTTP requests
type StreakHandler struct {
	streakService *services.StreakService
}

// NewStreakHandler creates a new streak handler
func NewStreakHandler(streakService *services.StreakService) *StreakHandler {
	return &StreakHandler{
		streakService: streakService,
	}
}

// UpdateStreakRequest is the request body for updating a streak
type UpdateStreakRequest struct {
	ProblemID    uint   `json:"problem_id" binding:"required"`
	SubmissionID string `json:"submission_id" binding:"required"`
}

// UpdateStreak handles POST /internal/streaks/update/:user_id
// Called after a successful submission to update the user's streak
func (h *StreakHandler) UpdateStreak(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	var req UpdateStreakRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	streak, err := h.streakService.UpdateStreak(userID, req.ProblemID, req.SubmissionID)
	if err != nil {
		if errors.Is(err, services.ErrNotDailyChallenge) {
			// Not an error - submission is not for daily challenge
			c.JSON(http.StatusOK, gin.H{
				"message":        "Not a daily challenge submission",
				"streak_updated": false,
			})
			return
		}
		if errors.Is(err, services.ErrAlreadySolved) {
			c.JSON(http.StatusOK, gin.H{
				"message":        "Daily challenge already solved today",
				"streak_updated": false,
				"streak":         streak,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update streak",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Streak updated successfully",
		"streak_updated": true,
		"streak":         streak,
	})
}

// GetStreak handles GET /internal/streaks/:user_id
func (h *StreakHandler) GetStreak(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	streak, err := h.streakService.GetStreak(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get streak",
		})
		return
	}

	c.JSON(http.StatusOK, streak)
}

// GetLeaderboard handles GET /internal/streaks/leaderboard
func (h *StreakHandler) GetLeaderboard(c *gin.Context) {
	limit := 10 // Default limit
	byLongest := c.Query("type") == "longest"

	streaks, err := h.streakService.GetLeaderboard(limit, byLongest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get leaderboard",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": streaks,
		"type":        map[bool]string{true: "longest", false: "current"}[byLongest],
	})
}
