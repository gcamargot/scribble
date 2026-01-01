package models

import (
	"time"
)

// UserAggregateMetrics represents computed statistics for a user
// Maps to the user_aggregate_metrics view in schema.sql
type UserAggregateMetrics struct {
	UserID                    uint       `gorm:"column:user_id" json:"user_id"`
	Username                  string     `json:"username"`
	AvatarURL                 *string    `gorm:"column:avatar_url" json:"avatar_url,omitempty"`
	ProblemsSolved            int        `gorm:"column:problems_solved" json:"problems_solved"`
	TotalAcceptedSubmissions  int        `gorm:"column:total_accepted_submissions" json:"total_accepted_submissions"`
	TotalSubmissions          int        `gorm:"column:total_submissions" json:"total_submissions"`
	AvgExecutionTimeMs        *float64   `gorm:"column:avg_execution_time_ms" json:"avg_execution_time_ms,omitempty"`
	AvgMemoryKb               *float64   `gorm:"column:avg_memory_kb" json:"avg_memory_kb,omitempty"`
	AcceptanceRate            *float64   `gorm:"column:acceptance_rate" json:"acceptance_rate,omitempty"`
	CurrentStreak             int        `gorm:"column:current_streak" json:"current_streak"`
	LongestStreak             int        `gorm:"column:longest_streak" json:"longest_streak"`
	LastSolvedDate            *time.Time `gorm:"column:last_solved_date" json:"last_solved_date,omitempty"`
	FavoriteLanguage          *string    `gorm:"column:favorite_language" json:"favorite_language,omitempty"`
}

// TableName specifies the view name
func (UserAggregateMetrics) TableName() string {
	return "user_aggregate_metrics"
}

// UserMetricsSummary is a compact version for leaderboards and lists
type UserMetricsSummary struct {
	UserID         uint    `json:"user_id"`
	Username       string  `json:"username"`
	AvatarURL      *string `json:"avatar_url,omitempty"`
	ProblemsSolved int     `json:"problems_solved"`
	CurrentStreak  int     `json:"current_streak"`
}

// ToSummary converts full metrics to a summary
func (m *UserAggregateMetrics) ToSummary() UserMetricsSummary {
	return UserMetricsSummary{
		UserID:         m.UserID,
		Username:       m.Username,
		AvatarURL:      m.AvatarURL,
		ProblemsSolved: m.ProblemsSolved,
		CurrentStreak:  m.CurrentStreak,
	}
}
