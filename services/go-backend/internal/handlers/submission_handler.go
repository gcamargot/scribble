package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nahtao97/scribble/internal/services"
)

// SubmissionHandler handles HTTP requests for submission-related endpoints
type SubmissionHandler struct {
	submissionService *services.SubmissionService
}

// NewSubmissionHandler creates a new submission handler instance
func NewSubmissionHandler(submissionService *services.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{
		submissionService: submissionService,
	}
}

// GetSubmissionByID handles GET /internal/submissions/:id
// Returns a specific submission by its ID
func (h *SubmissionHandler) GetSubmissionByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid submission ID",
		})
		return
	}

	// Check if code should be included
	includeCode := c.Query("include_code") == "true"

	if includeCode {
		submission, err := h.submissionService.GetSubmissionWithCode(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"submission": submission,
		})
		return
	}

	submission, err := h.submissionService.GetSubmissionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submission": submission,
	})
}

// GetUserSubmissionHistory handles GET /internal/submissions/user/:user_id
// Returns paginated submission history for a user
// Query params: page, page_size, problem_id, status, language
func (h *SubmissionHandler) GetUserSubmissionHistory(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Parse pagination parameters
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

	// Build params
	params := services.SubmissionHistoryParams{
		UserID:   uint(userID),
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
		Language: c.Query("language"),
	}

	// Parse optional problem ID filter
	if problemIDParam := c.Query("problem_id"); problemIDParam != "" {
		if pid, err := strconv.ParseUint(problemIDParam, 10, 32); err == nil {
			problemID := uint(pid)
			params.ProblemID = &problemID
		}
	}

	result, err := h.submissionService.GetUserSubmissionHistory(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUserSubmissionStats handles GET /internal/submissions/user/:user_id/stats
// Returns aggregate submission statistics for a user
func (h *SubmissionHandler) GetUserSubmissionStats(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	stats, err := h.submissionService.GetUserSubmissionStats(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
