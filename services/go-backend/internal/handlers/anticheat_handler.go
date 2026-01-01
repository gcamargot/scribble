package handlers

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nahtao97/scribble/internal/models"
	"github.com/nahtao97/scribble/internal/services"
)

// AntiCheatHandler handles HTTP requests for anti-cheat endpoints
type AntiCheatHandler struct {
	antiCheatService *services.AntiCheatService
}

// NewAntiCheatHandler creates a new anti-cheat handler instance
func NewAntiCheatHandler(antiCheatService *services.AntiCheatService) *AntiCheatHandler {
	return &AntiCheatHandler{
		antiCheatService: antiCheatService,
	}
}

// verifyAdmin checks if the request has valid admin credentials
func verifyAdmin(c *gin.Context) bool {
	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret == "" {
		return false
	}
	return c.GetHeader("X-Admin-Secret") == adminSecret
}

// GetPendingFlags handles GET /internal/admin/flags/pending
// Returns flagged submissions awaiting review (admin only)
func (h *AntiCheatHandler) GetPendingFlags(c *gin.Context) {
	if !verifyAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "admin access required"})
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 20
	if s := c.Query("page_size"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	flags, total, err := h.antiCheatService.GetPendingFlags(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"flags":       flags,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetFlagStats handles GET /internal/admin/flags/stats
// Returns aggregated statistics about flagged submissions (admin only)
func (h *AntiCheatHandler) GetFlagStats(c *gin.Context) {
	if !verifyAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "admin access required"})
		return
	}

	stats, err := h.antiCheatService.GetFlagStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// GetUserFlags handles GET /internal/admin/flags/user/:user_id
// Returns all flags for a specific user (admin only)
func (h *AntiCheatHandler) GetUserFlags(c *gin.Context) {
	if !verifyAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "admin access required"})
		return
	}

	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	flags, err := h.antiCheatService.GetFlagsByUser(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"flags":   flags,
		"count":   len(flags),
	})
}

// ReviewFlagRequest represents a flag review request
type ReviewFlagRequest struct {
	Status string `json:"status" binding:"required"`
	Notes  string `json:"notes"`
}

// ReviewFlag handles POST /internal/admin/flags/:flag_id/review
// Updates the status of a flagged submission (admin only)
func (h *AntiCheatHandler) ReviewFlag(c *gin.Context) {
	if !verifyAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "admin access required"})
		return
	}

	flagIDParam := c.Param("flag_id")
	flagID, err := strconv.ParseUint(flagIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flag ID"})
		return
	}

	var req ReviewFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate status
	status := models.FlagStatus(req.Status)
	validStatuses := []models.FlagStatus{
		models.FlagStatusReviewed,
		models.FlagStatusCleared,
		models.FlagStatusBanned,
	}

	valid := false
	for _, vs := range validStatuses {
		if status == vs {
			valid = true
			break
		}
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "invalid status",
			"valid_statuses": validStatuses,
		})
		return
	}

	// Get admin user ID from header (set by auth middleware)
	adminUserID := uint(0)
	if adminIDStr := c.GetHeader("X-User-Id"); adminIDStr != "" {
		if parsed, err := strconv.ParseUint(adminIDStr, 10, 32); err == nil {
			adminUserID = uint(parsed)
		}
	}

	err = h.antiCheatService.ReviewFlag(uint(flagID), adminUserID, status, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"flag_id": flagID,
		"status":  status,
	})
}

// CheckSubmissionRequest represents a submission check request
type CheckSubmissionRequest struct {
	UserID          uint   `json:"user_id" binding:"required"`
	ProblemID       uint   `json:"problem_id" binding:"required"`
	ExecutionTimeMs int    `json:"execution_time_ms"`
	MemoryUsedKb    int    `json:"memory_used_kb"`
	Difficulty      string `json:"difficulty"`
}

// CheckSubmission handles POST /internal/anticheat/check
// Checks if a submission should be allowed or flagged
func (h *AntiCheatHandler) CheckSubmission(c *gin.Context) {
	var req CheckSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.antiCheatService.CheckSubmission(
		req.UserID,
		req.ProblemID,
		req.ExecutionTimeMs,
		req.MemoryUsedKb,
		req.Difficulty,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// FlagSubmissionRequest represents a request to flag a submission
type FlagSubmissionRequest struct {
	SubmissionID uint                   `json:"submission_id" binding:"required"`
	UserID       uint                   `json:"user_id" binding:"required"`
	ProblemID    uint                   `json:"problem_id" binding:"required"`
	Reason       models.FlagReason      `json:"reason" binding:"required"`
	Details      map[string]interface{} `json:"details"`
}

// FlagSubmission handles POST /internal/anticheat/flag
// Creates a flag record for a submission
func (h *AntiCheatHandler) FlagSubmission(c *gin.Context) {
	var req FlagSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.antiCheatService.FlagSubmission(
		req.SubmissionID,
		req.UserID,
		req.ProblemID,
		req.Reason,
		req.Details,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"submission_id": req.SubmissionID,
		"reason":        req.Reason,
	})
}

// CleanupRateLimits handles POST /internal/admin/cleanup/rate-limits
// Removes stale rate limit entries (admin only, called by cron)
func (h *AntiCheatHandler) CleanupRateLimits(c *gin.Context) {
	if !verifyAdmin(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "admin access required"})
		return
	}

	deleted, err := h.antiCheatService.CleanupOldRateLimitEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"entries_deleted": deleted,
	})
}
