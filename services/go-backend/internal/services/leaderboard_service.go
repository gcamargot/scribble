package services

import (
	"fmt"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LeaderboardService handles leaderboard computation and retrieval
type LeaderboardService struct {
	db *gorm.DB
}

// NewLeaderboardService creates a new leaderboard service instance
func NewLeaderboardService(db *gorm.DB) *LeaderboardService {
	return &LeaderboardService{
		db: db,
	}
}

// ComputeAllLeaderboards computes rankings for all metric types
func (s *LeaderboardService) ComputeAllLeaderboards() ([]models.ComputeResult, error) {
	var results []models.ComputeResult

	for _, metricType := range models.AllMetricTypes() {
		result, err := s.ComputeLeaderboard(metricType)
		if err != nil {
			return results, fmt.Errorf("failed to compute %s: %w", metricType, err)
		}
		results = append(results, *result)
	}

	return results, nil
}

// ComputeLeaderboard computes rankings for a specific metric type
func (s *LeaderboardService) ComputeLeaderboard(metricType models.MetricType) (*models.ComputeResult, error) {
	now := time.Now()

	var entries []models.LeaderboardEntry
	var err error

	switch metricType {
	case models.MetricFastestAvg:
		entries, err = s.computeFastestAvg()
	case models.MetricLowestMemoryAvg:
		entries, err = s.computeLowestMemoryAvg()
	case models.MetricProblemsSolved:
		entries, err = s.computeProblemsSolved()
	case models.MetricLongestStreak:
		entries, err = s.computeLongestStreak()
	default:
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}

	if err != nil {
		return nil, err
	}

	// Upsert entries into leaderboard_cache
	if len(entries) > 0 {
		// Use ON CONFLICT to update existing entries
		result := s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "metric_type"}},
			DoUpdates: clause.AssignmentColumns([]string{"metric_value", "rank", "computed_at"}),
		}).Create(&entries)

		if result.Error != nil {
			return nil, fmt.Errorf("failed to upsert leaderboard entries: %w", result.Error)
		}
	}

	return &models.ComputeResult{
		MetricType:     metricType,
		EntriesUpdated: len(entries),
		ComputedAt:     now,
	}, nil
}

// computeFastestAvg calculates average execution time rankings
// Lower is better - ranks users by their average execution time for accepted submissions
func (s *LeaderboardService) computeFastestAvg() ([]models.LeaderboardEntry, error) {
	type userAvg struct {
		UserID uint
		Avg    float64
	}

	var results []userAvg

	// Calculate average execution time per user for accepted submissions
	err := s.db.Table("submissions").
		Select("user_id, AVG(execution_time_ms) as avg").
		Where("status = 'accepted' AND execution_time_ms IS NOT NULL").
		Group("user_id").
		Having("COUNT(*) >= 1"). // Require at least 1 accepted submission
		Order("avg ASC").        // Lower is better
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to compute fastest avg: %w", err)
	}

	entries := make([]models.LeaderboardEntry, len(results))
	for i, r := range results {
		entries[i] = models.LeaderboardEntry{
			UserID:      r.UserID,
			MetricType:  models.MetricFastestAvg,
			MetricValue: r.Avg,
			Rank:        i + 1,
			ComputedAt:  time.Now(),
		}
	}

	return entries, nil
}

// computeLowestMemoryAvg calculates average memory usage rankings
// Lower is better - ranks users by their average memory usage for accepted submissions
func (s *LeaderboardService) computeLowestMemoryAvg() ([]models.LeaderboardEntry, error) {
	type userAvg struct {
		UserID uint
		Avg    float64
	}

	var results []userAvg

	err := s.db.Table("submissions").
		Select("user_id, AVG(memory_used_kb) as avg").
		Where("status = 'accepted' AND memory_used_kb IS NOT NULL").
		Group("user_id").
		Having("COUNT(*) >= 1").
		Order("avg ASC"). // Lower is better
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to compute lowest memory avg: %w", err)
	}

	entries := make([]models.LeaderboardEntry, len(results))
	for i, r := range results {
		entries[i] = models.LeaderboardEntry{
			UserID:      r.UserID,
			MetricType:  models.MetricLowestMemoryAvg,
			MetricValue: r.Avg,
			Rank:        i + 1,
			ComputedAt:  time.Now(),
		}
	}

	return entries, nil
}

// computeProblemsSolved calculates unique problems solved rankings
// Higher is better - counts distinct problems with at least one accepted submission
func (s *LeaderboardService) computeProblemsSolved() ([]models.LeaderboardEntry, error) {
	type userCount struct {
		UserID uint
		Count  int
	}

	var results []userCount

	err := s.db.Table("submissions").
		Select("user_id, COUNT(DISTINCT problem_id) as count").
		Where("status = 'accepted'").
		Group("user_id").
		Order("count DESC"). // Higher is better
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to compute problems solved: %w", err)
	}

	entries := make([]models.LeaderboardEntry, len(results))
	for i, r := range results {
		entries[i] = models.LeaderboardEntry{
			UserID:      r.UserID,
			MetricType:  models.MetricProblemsSolved,
			MetricValue: float64(r.Count),
			Rank:        i + 1,
			ComputedAt:  time.Now(),
		}
	}

	return entries, nil
}

// computeLongestStreak calculates longest streak rankings
// Higher is better - based on the streaks table
func (s *LeaderboardService) computeLongestStreak() ([]models.LeaderboardEntry, error) {
	type userStreak struct {
		UserID        uint
		LongestStreak int
	}

	var results []userStreak

	err := s.db.Table("streaks").
		Select("user_id, longest_streak").
		Where("longest_streak > 0").
		Order("longest_streak DESC"). // Higher is better
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to compute longest streak: %w", err)
	}

	entries := make([]models.LeaderboardEntry, len(results))
	for i, r := range results {
		entries[i] = models.LeaderboardEntry{
			UserID:      r.UserID,
			MetricType:  models.MetricLongestStreak,
			MetricValue: float64(r.LongestStreak),
			Rank:        i + 1,
			ComputedAt:  time.Now(),
		}
	}

	return entries, nil
}

// GetLeaderboard retrieves paginated leaderboard for a metric type
func (s *LeaderboardService) GetLeaderboard(metricType models.MetricType, page, pageSize int) (*models.LeaderboardPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get total count
	var total int64
	s.db.Model(&models.LeaderboardEntry{}).
		Where("metric_type = ?", metricType).
		Count(&total)

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	offset := (page - 1) * pageSize

	// Fetch entries with user info
	var entries []models.LeaderboardWithUser

	err := s.db.Table("leaderboard_cache lc").
		Select("lc.*, u.username, u.avatar_url").
		Joins("JOIN users u ON lc.user_id = u.id").
		Where("lc.metric_type = ?", metricType).
		Order("lc.rank ASC").
		Offset(offset).
		Limit(pageSize).
		Scan(&entries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return &models.LeaderboardPage{
		Entries:    entries,
		MetricType: metricType,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Total:      total,
	}, nil
}

// GetUserRank retrieves a user's rank for a specific metric
func (s *LeaderboardService) GetUserRank(userID uint, metricType models.MetricType) (*models.LeaderboardEntry, error) {
	var entry models.LeaderboardEntry

	result := s.db.Where("user_id = ? AND metric_type = ?", userID, metricType).First(&entry)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // User not ranked yet
		}
		return nil, fmt.Errorf("failed to get user rank: %w", result.Error)
	}

	return &entry, nil
}

// GetUserAllRanks retrieves a user's ranks for all metrics
func (s *LeaderboardService) GetUserAllRanks(userID uint) ([]models.LeaderboardEntry, error) {
	var entries []models.LeaderboardEntry

	result := s.db.Where("user_id = ?", userID).Find(&entries)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user ranks: %w", result.Error)
	}

	return entries, nil
}
