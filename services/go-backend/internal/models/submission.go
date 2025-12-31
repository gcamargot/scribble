package models

import (
	"time"
)

// SubmissionStatus represents the result status of a code submission
type SubmissionStatus string

const (
	StatusAccepted         SubmissionStatus = "accepted"
	StatusWrongAnswer      SubmissionStatus = "wrong_answer"
	StatusRuntimeError     SubmissionStatus = "runtime_error"
	StatusTimeout          SubmissionStatus = "timeout"
	StatusMemoryLimit      SubmissionStatus = "memory_limit"
	StatusCompilationError SubmissionStatus = "compilation_error"
)

// Submission represents a user's code submission
// Maps to the submissions table in schema.sql
type Submission struct {
	ID               uint             `gorm:"primaryKey" json:"id"`
	UserID           uint             `gorm:"not null;index" json:"user_id"`
	ProblemID        uint             `gorm:"not null;index" json:"problem_id"`
	DailyChallengeID *uint            `gorm:"index" json:"daily_challenge_id,omitempty"`
	Language         string           `gorm:"not null" json:"language"`
	Code             string           `gorm:"not null" json:"-"` // Don't expose code in responses
	Status           SubmissionStatus `gorm:"not null" json:"status"`

	// Execution metrics
	ExecutionTimeMs *int `json:"execution_time_ms,omitempty"`
	MemoryUsedKb    *int `json:"memory_used_kb,omitempty"`
	TestsPassed     int  `gorm:"default:0" json:"tests_passed"`
	TestsTotal      int  `gorm:"default:0" json:"tests_total"`

	// Metadata
	SubmittedAt  time.Time `gorm:"autoCreateTime" json:"submitted_at"`
	ErrorMessage *string   `json:"error_message,omitempty"`

	// Associations
	Problem        Problem         `gorm:"foreignKey:ProblemID" json:"problem,omitempty"`
	DailyChallenge *DailyChallenge `gorm:"foreignKey:DailyChallengeID" json:"daily_challenge,omitempty"`
}

// TableName specifies the table name for the Submission model
func (Submission) TableName() string {
	return "submissions"
}

// PassRate returns the percentage of tests passed (0-100)
func (s *Submission) PassRate() float64 {
	if s.TestsTotal == 0 {
		return 0
	}
	return float64(s.TestsPassed) / float64(s.TestsTotal) * 100
}

// IsAccepted returns true if the submission passed all tests
func (s *Submission) IsAccepted() bool {
	return s.Status == StatusAccepted
}

// PercentileMetrics contains computed percentile comparisons for a submission
type PercentileMetrics struct {
	SubmissionID uint `json:"submission_id"`
	ProblemID    uint `json:"problem_id"`

	// Execution time percentile (lower time = higher percentile)
	// e.g., 78 means "Faster than 78% of submissions"
	ExecutionTimePercentile *float64 `json:"execution_time_percentile,omitempty"`
	ExecutionTimeRank       *int     `json:"execution_time_rank,omitempty"`
	TotalSubmissions        int      `json:"total_submissions"`

	// Memory usage percentile (lower memory = higher percentile)
	// e.g., 65 means "Uses less memory than 65% of submissions"
	MemoryPercentile *float64 `json:"memory_percentile,omitempty"`
	MemoryRank       *int     `json:"memory_rank,omitempty"`

	// Human-readable messages
	ExecutionTimeMessage string `json:"execution_time_message,omitempty"`
	MemoryMessage        string `json:"memory_message,omitempty"`
}
