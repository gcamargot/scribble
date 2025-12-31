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

// GetPercentileMetrics handles GET /internal/submissions/:id/percentile
// Returns percentile comparison metrics for a submission
// e.g., "Faster than 78% of submissions", "Uses less memory than 65% of submissions"
func (h *SubmissionHandler) GetPercentileMetrics(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid submission ID",
		})
		return
	}

	metrics, err := h.submissionService.CalculatePercentileMetrics(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"percentile": metrics,
	})
}

// GetProblemStats handles GET /internal/problems/:id/stats
// Returns aggregate submission statistics for a problem
func (h *SubmissionHandler) GetProblemStats(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid problem ID",
		})
		return
	}

	stats, err := h.submissionService.GetProblemSubmissionStats(uint(id))
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
