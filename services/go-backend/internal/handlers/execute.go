package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nahtao97/scribble/internal/k8s"
	"github.com/nahtao97/scribble/internal/models"
)

// ExecuteRequest represents the request body for code execution
type ExecuteRequest struct {
	Code      string `json:"code" binding:"required"`
	Language  string `json:"language" binding:"required"`
	ProblemID string `json:"problem_id" binding:"required"`
	UserID    string `json:"user_id" binding:"required"`
}

// ExecuteResponse represents the response from code execution
type ExecuteResponse struct {
	SubmissionID    string              `json:"submission_id"`
	Status          string              `json:"status"`
	ExecutionTimeMs int64               `json:"execution_time_ms"`
	MemoryUsedKB    int64               `json:"memory_used_kb"`
	TestsPassed     int                 `json:"tests_passed"`
	TestsTotal      int                 `json:"tests_total"`
	ErrorMessage    string              `json:"error_message,omitempty"`
	TestResults     []k8s.TestResult    `json:"test_results,omitempty"`
}

// ExecuteHandler handles code execution requests
type ExecuteHandler struct {
	jobManager *k8s.JobManager
	// TODO: Add database connection for submission storage
	// TODO: Add problem service for test case retrieval
}

// NewExecuteHandler creates a new execute handler
func NewExecuteHandler(jobManager *k8s.JobManager) *ExecuteHandler {
	return &ExecuteHandler{
		jobManager: jobManager,
	}
}

// Execute handles POST /internal/execute
// It validates the request, creates a submission, executes code, and returns results
func (h *ExecuteHandler) Execute(c *gin.Context) {
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate language
	if !models.IsValidLanguage(req.Language) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported language",
			"supported": models.ValidLanguages,
		})
		return
	}

	// TODO: Create submission record in database
	// For now, generate a simple ID
	submissionID := generateSubmissionID()

	// TODO: Retrieve test cases from database based on problem_id
	// For now, use placeholder test cases
	testCases := getPlaceholderTestCases(req.ProblemID)

	// Execute code using K8s job
	result, err := h.jobManager.ExecuteAndWait(c.Request.Context(), k8s.ExecutionJobParams{
		SubmissionID: submissionID,
		ProblemID:    req.ProblemID,
		Code:         req.Code,
		TestCases:    testCases,
	})

	if err != nil {
		// Even on error, we may have a partial result
		if result == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Execution failed",
				"details": err.Error(),
			})
			return
		}
	}

	// TODO: Update submission record in database with results

	// Return execution result
	c.JSON(http.StatusOK, ExecuteResponse{
		SubmissionID:    submissionID,
		Status:          result.Status,
		ExecutionTimeMs: result.ExecutionTimeMs,
		MemoryUsedKB:    result.MemoryUsedKB,
		TestsPassed:     result.TestsPassed,
		TestsTotal:      result.TestsTotal,
		ErrorMessage:    result.ErrorMessage,
		TestResults:     result.TestResults,
	})
}

// generateSubmissionID creates a unique submission ID
// TODO: Replace with UUID from database
func generateSubmissionID() string {
	return "sub-" + randomString(12)
}

// randomString generates a random alphanumeric string
func randomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[i%len(chars)]
	}
	return string(b)
}

// getPlaceholderTestCases returns test cases for a problem
// TODO: Replace with database lookup
func getPlaceholderTestCases(problemID string) []map[string]interface{} {
	// Placeholder test cases for "two-sum" problem
	return []map[string]interface{}{
		{
			"input": map[string]interface{}{
				"nums":   []int{2, 7, 11, 15},
				"target": 9,
			},
			"expected_output": []int{0, 1},
		},
		{
			"input": map[string]interface{}{
				"nums":   []int{3, 2, 4},
				"target": 6,
			},
			"expected_output": []int{1, 2},
		},
	}
}
