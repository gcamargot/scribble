package models

import (
	"time"

	"gorm.io/gorm"
)

// Problem represents a coding challenge in the database
// Maps to the problems table in schema.sql
type Problem struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	Slug        string         `gorm:"uniqueIndex;not null" json:"slug"`
	Difficulty  string         `gorm:"not null" json:"difficulty"` // 'easy', 'medium', 'hard'
	Description string         `gorm:"not null" json:"description"`
	Constraints string         `json:"constraints,omitempty"`
	Hints       []string       `gorm:"type:text[]" json:"hints,omitempty"`
	Category    string         `json:"category,omitempty"`
	Tags        []string       `gorm:"type:text[]" json:"tags,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Problem model
func (Problem) TableName() string {
	return "problems"
}

// TestCase represents test input/output for a problem
// Maps to the test_cases table in schema.sql
type TestCase struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ProblemID      uint           `gorm:"not null;index" json:"problem_id"`
	Input          interface{}    `gorm:"type:jsonb;not null" json:"input"`           // JSONB input data
	ExpectedOutput interface{}    `gorm:"type:jsonb;not null" json:"expected_output"` // JSONB expected output
	IsSample       bool           `gorm:"default:false" json:"is_sample"`             // true if this is a sample test case shown to users
	Weight         float64        `gorm:"default:1.0" json:"weight,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the TestCase model
func (TestCase) TableName() string {
	return "test_cases"
}

// DailyChallenge represents the daily coding challenge
// Maps to the daily_challenges table in schema.sql
type DailyChallenge struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProblemID     uint      `gorm:"not null;index" json:"problem_id"`
	ChallengeDate time.Time `gorm:"uniqueIndex;not null;type:date" json:"challenge_date"` // Date of the challenge (UTC)
	CreatedAt     time.Time `json:"created_at"`

	// Associations - GORM will handle foreign key relationships
	Problem Problem `gorm:"foreignKey:ProblemID" json:"problem,omitempty"`
}

// TableName specifies the table name for the DailyChallenge model
func (DailyChallenge) TableName() string {
	return "daily_challenges"
}
