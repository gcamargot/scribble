package models

import (
	"time"
)

// Streak represents a user's streak record for the daily challenge
type Streak struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	CurrentStreak  int        `gorm:"default:0" json:"current_streak"`
	LongestStreak  int        `gorm:"default:0" json:"longest_streak"`
	LastSolvedDate *time.Time `gorm:"type:date" json:"last_solved_date"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for GORM
func (Streak) TableName() string {
	return "streaks"
}

// StreakUpdate represents the result of a streak update operation
type StreakUpdate struct {
	UserID         uint      `json:"user_id"`
	PreviousStreak int       `json:"previous_streak"`
	CurrentStreak  int       `json:"current_streak"`
	LongestStreak  int       `json:"longest_streak"`
	IsNewRecord    bool      `json:"is_new_record"`
	WasReset       bool      `json:"was_reset"`
	SolvedDate     time.Time `json:"solved_date"`
}

// StreakStats represents aggregated streak statistics
type StreakStats struct {
	TotalUsers         int64   `json:"total_users"`
	AverageStreak      float64 `json:"average_streak"`
	MaxActiveStreak    int     `json:"max_active_streak"`
	UsersWithStreak    int64   `json:"users_with_streak"`
	StreakDistribution map[string]int64 `json:"streak_distribution"`
}
