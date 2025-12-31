package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nahtao97/scribble/internal/services"
)

// ProblemHandler handles HTTP requests for problem-related endpoints
type ProblemHandler struct {
	problemService *services.ProblemService
}

// NewProblemHandler creates a new problem handler instance
func NewProblemHandler(problemService *services.ProblemService) *ProblemHandler {
	return &ProblemHandler{
		problemService: problemService,
	}
}

// GetProblemByID handles GET /internal/problems/:id
// Returns a specific problem by its ID
func (h *ProblemHandler) GetProblemByID(c *gin.Context) {
	// Parse problem ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid problem ID",
		})
		return
	}

	// Retrieve problem from service
	problem, err := h.problemService.GetProblemByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return problem data
	c.JSON(http.StatusOK, gin.H{
		"problem": problem,
	})
}

// GetTestCasesByProblemID handles GET /internal/problems/:id/test-cases
// Returns test cases for a specific problem (sample tests only by default)
// Query param: all=true to include hidden tests (for internal executor use)
func (h *ProblemHandler) GetTestCasesByProblemID(c *gin.Context) {
	// Parse problem ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid problem ID",
		})
		return
	}

	// Check if all tests should be included (internal use only)
	// By default, only return sample tests to prevent cheating
	sampleOnly := true
	if c.Query("all") == "true" {
		sampleOnly = false
	}

	// Retrieve test cases from service
	testCases, err := h.problemService.GetTestCasesByProblemID(uint(id), sampleOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return test cases
	c.JSON(http.StatusOK, gin.H{
		"test_cases":  testCases,
		"total":       len(testCases),
		"sample_only": sampleOnly,
	})
}

// GetDailyChallengeByDate handles GET /internal/problems/daily/:date
// Returns the daily challenge for a specific date
// Date format: YYYY-MM-DD (e.g., 2025-12-31)
// If :date is "today", returns today's challenge
func (h *ProblemHandler) GetDailyChallengeByDate(c *gin.Context) {
	// Parse date from URL parameter
	dateParam := c.Param("date")

	var challenge interface{}
	var err error

	// Handle "today" as a special case
	if dateParam == "today" {
		challenge, err = h.problemService.GetTodaysDailyChallenge()
	} else {
		// Parse date in YYYY-MM-DD format
		date, parseErr := time.Parse("2006-01-02", dateParam)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid date format, use YYYY-MM-DD or 'today'",
			})
			return
		}
		challenge, err = h.problemService.GetDailyChallengeByDate(date)
	}

	// Handle errors from service
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return daily challenge with associated problem
	c.JSON(http.StatusOK, gin.H{
		"daily_challenge": challenge,
	})
}
