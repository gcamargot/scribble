package models

import (
	"time"

	"gorm.io/gorm"
)

// Submission status constants
const (
	StatusPending          = "pending"
	StatusRunning          = "running"
	StatusAccepted         = "accepted"
	StatusWrongAnswer      = "wrong_answer"
	StatusRuntimeError     = "runtime_error"
	StatusTimeLimit        = "time_limit"
	StatusMemoryLimit      = "memory_limit"
	StatusCompilationError = "compilation_error"
)

// Submission represents a user's code submission
type Submission struct {
	ID        string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string         `json:"user_id" gorm:"not null;index"`
	ProblemID string         `json:"problem_id" gorm:"not null;index"`
	Language  string         `json:"language" gorm:"not null"`
	Code      string         `json:"code" gorm:"type:text;not null"`
	Status    string         `json:"status" gorm:"not null;default:'pending'"`

	// Time breakdown fields
	CompilationTimeMs   int `json:"compilation_time_ms" gorm:"default:0"`
	ExecutionTimeMs     int `json:"execution_time_ms" gorm:"default:0"`      // Average per test
	TotalExecutionTimeMs int `json:"total_execution_time_ms" gorm:"default:0"` // Sum of all tests

	// Memory usage
	MemoryUsedKb int `json:"memory_used_kb" gorm:"default:0"`

	// Test results
	TestsPassed int `json:"tests_passed" gorm:"default:0"`
	TestsTotal  int `json:"tests_total" gorm:"default:0"`

	// Error details (for failed submissions)
	ErrorMessage string `json:"error_message,omitempty" gorm:"type:text"`
	ErrorType    string `json:"error_type,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for the Submission model
func (Submission) TableName() string {
	return "submissions"
}

// TestResult represents the result of a single test case execution
type TestResult struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	SubmissionID string `json:"submission_id" gorm:"not null;index;type:uuid"`
	TestCaseID   int    `json:"test_case_id" gorm:"not null"`
	Passed       bool   `json:"passed" gorm:"not null"`

	// Per-test timing
	ExecutionTimeMs int `json:"execution_time_ms" gorm:"default:0"`

	// Output comparison
	ActualOutput   string `json:"actual_output,omitempty" gorm:"type:text"`
	ExpectedOutput string `json:"expected_output,omitempty" gorm:"type:text"`

	// Error details (if this test failed with error)
	ErrorMessage string `json:"error_message,omitempty" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for the TestResult model
func (TestResult) TableName() string {
	return "test_results"
}

// SubmissionWithDetails includes the test results
type SubmissionWithDetails struct {
	Submission
	TestResults []TestResult `json:"test_results,omitempty" gorm:"foreignKey:SubmissionID"`
}

// PassRate calculates the percentage of tests passed
func (s *Submission) PassRate() float64 {
	if s.TestsTotal == 0 {
		return 0
	}
	return float64(s.TestsPassed) / float64(s.TestsTotal) * 100
}

// IsComplete returns true if the submission has finished processing
func (s *Submission) IsComplete() bool {
	return s.Status != StatusPending && s.Status != StatusRunning
}

// IsSuccessful returns true if all tests passed
func (s *Submission) IsSuccessful() bool {
	return s.Status == StatusAccepted
}
