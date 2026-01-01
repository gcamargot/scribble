package models

import (
	"time"
)

// MetricType represents the type of leaderboard metric
type MetricType string

const (
	MetricFastestAvg      MetricType = "fastest_avg"
	MetricLowestMemoryAvg MetricType = "lowest_memory_avg"
	MetricProblemsSolved  MetricType = "problems_solved"
	MetricLongestStreak   MetricType = "longest_streak"
)

// AllMetricTypes returns all available metric types
func AllMetricTypes() []MetricType {
	return []MetricType{
		MetricFastestAvg,
		MetricLowestMemoryAvg,
		MetricProblemsSolved,
		MetricLongestStreak,
	}
}

// LeaderboardEntry represents a cached leaderboard entry
// Maps to leaderboard_cache table in schema.sql
type LeaderboardEntry struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"not null;uniqueIndex:idx_user_metric" json:"user_id"`
	MetricType  MetricType `gorm:"not null;uniqueIndex:idx_user_metric" json:"metric_type"`
	MetricValue float64    `gorm:"type:decimal(12,2);not null" json:"metric_value"`
	Rank        int        `gorm:"not null" json:"rank"`
	ComputedAt  time.Time  `gorm:"autoCreateTime" json:"computed_at"`
}

// TableName specifies the table name for LeaderboardEntry
func (LeaderboardEntry) TableName() string {
	return "leaderboard_cache"
}

// LeaderboardWithUser extends LeaderboardEntry with user info
type LeaderboardWithUser struct {
	LeaderboardEntry
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// ComputeResult contains statistics about a leaderboard computation
type ComputeResult struct {
	MetricType     MetricType `json:"metric_type"`
	EntriesUpdated int        `json:"entries_updated"`
	ComputedAt     time.Time  `json:"computed_at"`
}

// LeaderboardPage represents a paginated leaderboard response
type LeaderboardPage struct {
	Entries    []LeaderboardWithUser `json:"entries"`
	MetricType MetricType            `json:"metric_type"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
	Total      int64                 `json:"total"`
}
