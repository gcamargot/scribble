package models

import (
	"time"
)

// FlagReason represents why a submission was flagged
type FlagReason string

const (
	FlagReasonSuspiciousTime   FlagReason = "suspicious_time"   // Execution time too fast for problem difficulty
	FlagReasonZeroMemory       FlagReason = "zero_memory"       // 0 KB memory usage (impossible)
	FlagReasonRateLimitAbuse   FlagReason = "rate_limit_abuse"  // Too many submissions in short period
	FlagReasonIdenticalCode    FlagReason = "identical_code"    // Same code submitted by multiple users
	FlagReasonPatternMatch     FlagReason = "pattern_match"     // Known cheating pattern detected
)

// FlagStatus represents the review status of a flagged submission
type FlagStatus string

const (
	FlagStatusPending  FlagStatus = "pending"  // Awaiting admin review
	FlagStatusReviewed FlagStatus = "reviewed" // Admin has reviewed
	FlagStatusCleared  FlagStatus = "cleared"  // False positive, submission is legitimate
	FlagStatusBanned   FlagStatus = "banned"   // Confirmed cheating, user may be banned
)

// FlaggedSubmission represents a submission flagged for potential cheating
type FlaggedSubmission struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	SubmissionID uint       `gorm:"not null;index" json:"submission_id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	ProblemID    uint       `gorm:"not null" json:"problem_id"`
	Reason       FlagReason `gorm:"not null" json:"reason"`
	Details      string     `gorm:"type:text" json:"details,omitempty"` // JSON details about the flag
	Status       FlagStatus `gorm:"default:pending" json:"status"`
	ReviewedBy   *uint      `json:"reviewed_by,omitempty"`   // Admin user ID who reviewed
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes  *string    `gorm:"type:text" json:"review_notes,omitempty"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (FlaggedSubmission) TableName() string {
	return "flagged_submissions"
}

// RateLimitEntry tracks submission rate for a user
type RateLimitEntry struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;uniqueIndex" json:"user_id"`
	Submissions int       `gorm:"default:0" json:"submissions"`   // Count in current window
	WindowStart time.Time `gorm:"not null" json:"window_start"`   // Start of rate limit window
	LastSubmit  time.Time `gorm:"not null" json:"last_submit"`    // Last submission time
}

// TableName specifies the table name
func (RateLimitEntry) TableName() string {
	return "rate_limit_entries"
}

// SuspiciousTimeThresholds defines minimum expected execution times by difficulty
// Submissions faster than these are flagged as suspicious
var SuspiciousTimeThresholds = map[string]int{
	"easy":   5,   // 5ms minimum for easy problems
	"medium": 10,  // 10ms minimum for medium problems
	"hard":   20,  // 20ms minimum for hard problems
}

// RateLimitConfig defines rate limiting parameters
type RateLimitConfig struct {
	WindowDuration   time.Duration // Time window for counting submissions
	MaxSubmissions   int           // Max submissions allowed in window
	CooldownDuration time.Duration // Cooldown after hitting limit
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		WindowDuration:   time.Minute * 5,  // 5-minute window
		MaxSubmissions:   10,               // 10 submissions per window
		CooldownDuration: time.Minute * 10, // 10-minute cooldown
	}
}
