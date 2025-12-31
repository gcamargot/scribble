package models

import (
	"time"

	"gorm.io/gorm"
)

// UserStreak tracks a user's daily challenge streak
type UserStreak struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	UserID          string         `json:"user_id" gorm:"uniqueIndex;not null"`
	CurrentStreak   int            `json:"current_streak" gorm:"not null;default:0"`
	LongestStreak   int            `json:"longest_streak" gorm:"not null;default:0"`
	LastSolvedDate  *time.Time     `json:"last_solved_date" gorm:"type:date"`
	TotalDaysSolved int            `json:"total_days_solved" gorm:"not null;default:0"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for the UserStreak model
func (UserStreak) TableName() string {
	return "user_streaks"
}

// StreakHistory tracks historical streak data for analytics
type StreakHistory struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"not null;index"`
	SolvedDate   time.Time `json:"solved_date" gorm:"type:date;not null"`
	ProblemID    uint      `json:"problem_id" gorm:"not null"`
	SubmissionID string    `json:"submission_id" gorm:"type:uuid"`
	StreakDay    int       `json:"streak_day" gorm:"not null"` // Day number in the streak
	CreatedAt    time.Time `json:"created_at"`
}

// TableName specifies the table name for the StreakHistory model
func (StreakHistory) TableName() string {
	return "streak_history"
}
