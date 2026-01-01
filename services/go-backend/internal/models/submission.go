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

// AllStatuses returns all valid submission statuses
func AllStatuses() []SubmissionStatus {
	return []SubmissionStatus{
		StatusAccepted,
		StatusWrongAnswer,
		StatusRuntimeError,
		StatusTimeout,
		StatusMemoryLimit,
		StatusCompilationError,
	}
}

// IsValidStatus checks if a status string is valid
func IsValidStatus(status string) bool {
	for _, s := range AllStatuses() {
		if string(s) == status {
			return true
		}
	}
	return false
}

// SupportedLanguages lists the supported programming languages
var SupportedLanguages = []string{"python", "javascript", "go", "java", "cpp", "rust"}

// IsValidLanguage checks if a language is supported
func IsValidLanguage(lang string) bool {
	for _, l := range SupportedLanguages {
		if l == lang {
			return true
		}
	}
	return false
}

// Submission represents a user's code submission
// Maps to the submissions table in schema.sql
type Submission struct {
	ID               uint             `gorm:"primaryKey" json:"id"`
	UserID           uint             `gorm:"not null;index" json:"user_id"`
	ProblemID        uint             `gorm:"not null;index" json:"problem_id"`
	DailyChallengeID *uint            `gorm:"index" json:"daily_challenge_id,omitempty"`
	Language         string           `gorm:"not null" json:"language"`
	Code             string           `gorm:"not null" json:"-"` // Don't expose code in list responses
	Status           SubmissionStatus `gorm:"not null" json:"status"`

	// Execution metrics
	ExecutionTimeMs *int `json:"execution_time_ms,omitempty"`
	MemoryUsedKb    *int `json:"memory_used_kb,omitempty"`
	TestsPassed     int  `gorm:"default:0" json:"tests_passed"`
	TestsTotal      int  `gorm:"default:0" json:"tests_total"`

	// Metadata
	SubmittedAt  time.Time `gorm:"autoCreateTime" json:"submitted_at"`
	ErrorMessage *string   `json:"error_message,omitempty"`

	// Associations (loaded via Preload)
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

// SubmissionWithCode includes the code field for detailed view
type SubmissionWithCode struct {
	Submission
	Code string `json:"code"`
}

// ToWithCode creates a SubmissionWithCode from a Submission
func (s *Submission) ToWithCode(code string) *SubmissionWithCode {
	return &SubmissionWithCode{
		Submission: *s,
		Code:       code,
	}
}
