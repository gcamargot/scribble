package services

import (
	"fmt"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/gorm"
)

// UserService handles user-related operations
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service instance
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// GetUserMetrics retrieves aggregate metrics for a specific user
func (s *UserService) GetUserMetrics(userID uint) (*models.UserAggregateMetrics, error) {
	var metrics models.UserAggregateMetrics

	result := s.db.Where("user_id = ?", userID).First(&metrics)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with ID %d not found", userID)
		}
		return nil, fmt.Errorf("failed to retrieve user metrics: %w", result.Error)
	}

	return &metrics, nil
}

// GetUserMetricsByUsername retrieves aggregate metrics by username
func (s *UserService) GetUserMetricsByUsername(username string) (*models.UserAggregateMetrics, error) {
	var metrics models.UserAggregateMetrics

	result := s.db.Where("username = ?", username).First(&metrics)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user '%s' not found", username)
		}
		return nil, fmt.Errorf("failed to retrieve user metrics: %w", result.Error)
	}

	return &metrics, nil
}

// GetTopUsersByProblems retrieves users with most problems solved
func (s *UserService) GetTopUsersByProblems(limit int) ([]models.UserMetricsSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var metrics []models.UserAggregateMetrics

	result := s.db.Order("problems_solved DESC").Limit(limit).Find(&metrics)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve top users: %w", result.Error)
	}

	summaries := make([]models.UserMetricsSummary, len(metrics))
	for i, m := range metrics {
		summaries[i] = m.ToSummary()
	}

	return summaries, nil
}

// GetTopUsersByStreak retrieves users with longest current streaks
func (s *UserService) GetTopUsersByStreak(limit int) ([]models.UserMetricsSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var metrics []models.UserAggregateMetrics

	result := s.db.Order("current_streak DESC").Limit(limit).Find(&metrics)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve top streaks: %w", result.Error)
	}

	summaries := make([]models.UserMetricsSummary, len(metrics))
	for i, m := range metrics {
		summaries[i] = m.ToSummary()
	}

	return summaries, nil
}

// GetLanguageStats returns language usage statistics for a user
func (s *UserService) GetLanguageStats(userID uint) ([]LanguageStat, error) {
	var stats []LanguageStat

	result := s.db.Table("submissions").
		Select("language, COUNT(*) as count, COUNT(*) FILTER (WHERE status = 'accepted') as accepted_count").
		Where("user_id = ?", userID).
		Group("language").
		Order("count DESC").
		Scan(&stats)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve language stats: %w", result.Error)
	}

	return stats, nil
}

// LanguageStat represents language usage statistics
type LanguageStat struct {
	Language      string `json:"language"`
	Count         int    `json:"count"`
	AcceptedCount int    `json:"accepted_count"`
}
