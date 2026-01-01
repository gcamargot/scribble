package services

import (
	"fmt"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StreakService handles user streak calculations with timezone awareness
type StreakService struct {
	db *gorm.DB
	// defaultTimezone is used when user timezone is not specified
	// All streak calculations are normalized to this timezone
	defaultTimezone *time.Location
}

// NewStreakService creates a new streak service instance
func NewStreakService(db *gorm.DB) *StreakService {
	// Use UTC as default timezone for consistency
	return &StreakService{
		db:              db,
		defaultTimezone: time.UTC,
	}
}

// NewStreakServiceWithTimezone creates a streak service with a custom default timezone
func NewStreakServiceWithTimezone(db *gorm.DB, timezone *time.Location) *StreakService {
	return &StreakService{
		db:              db,
		defaultTimezone: timezone,
	}
}

// GetUserStreak retrieves the current streak for a user
func (s *StreakService) GetUserStreak(userID uint) (*models.Streak, error) {
	var streak models.Streak
	result := s.db.Where("user_id = ?", userID).First(&streak)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user streak: %w", result.Error)
	}
	return &streak, nil
}

// UpdateStreak updates a user's streak after solving the daily problem
// userTimezone is the user's local timezone (e.g., "America/New_York")
// If empty, uses the default timezone (UTC)
func (s *StreakService) UpdateStreak(userID uint, userTimezone string) (*models.StreakUpdate, error) {
	// Determine the timezone to use
	loc := s.defaultTimezone
	if userTimezone != "" {
		parsed, err := time.LoadLocation(userTimezone)
		if err == nil {
			loc = parsed
		}
		// If timezone is invalid, fall back to default
	}

	// Get current time in user's timezone
	now := time.Now().In(loc)
	today := s.truncateToDate(now, loc)

	// Get or create streak record
	var streak models.Streak
	result := s.db.Where("user_id = ?", userID).First(&streak)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get streak: %w", result.Error)
	}

	isNew := result.Error == gorm.ErrRecordNotFound
	if isNew {
		streak = models.Streak{
			UserID:        userID,
			CurrentStreak: 0,
			LongestStreak: 0,
		}
	}

	update := &models.StreakUpdate{
		UserID:         userID,
		PreviousStreak: streak.CurrentStreak,
		SolvedDate:     today,
	}

	// Check if already solved today
	if streak.LastSolvedDate != nil {
		lastSolved := s.truncateToDate(*streak.LastSolvedDate, loc)
		if lastSolved.Equal(today) {
			// Already solved today, no change
			update.CurrentStreak = streak.CurrentStreak
			update.LongestStreak = streak.LongestStreak
			return update, nil
		}

		// Check if solved yesterday (consecutive day)
		yesterday := today.AddDate(0, 0, -1)
		if lastSolved.Equal(yesterday) {
			// Continue streak
			streak.CurrentStreak++
		} else {
			// Streak broken - reset to 1
			update.WasReset = true
			streak.CurrentStreak = 1
		}
	} else {
		// First time solving
		streak.CurrentStreak = 1
	}

	// Update longest streak if needed
	if streak.CurrentStreak > streak.LongestStreak {
		streak.LongestStreak = streak.CurrentStreak
		update.IsNewRecord = true
	}

	streak.LastSolvedDate = &today
	update.CurrentStreak = streak.CurrentStreak
	update.LongestStreak = streak.LongestStreak

	// Save the streak
	if isNew {
		if err := s.db.Create(&streak).Error; err != nil {
			return nil, fmt.Errorf("failed to create streak: %w", err)
		}
	} else {
		if err := s.db.Save(&streak).Error; err != nil {
			return nil, fmt.Errorf("failed to update streak: %w", err)
		}
	}

	return update, nil
}

// CheckStreak checks if a user's streak is still valid (not broken)
// This should be called to update streak status without counting a new solve
func (s *StreakService) CheckStreak(userID uint, userTimezone string) (*models.Streak, bool, error) {
	loc := s.defaultTimezone
	if userTimezone != "" {
		parsed, err := time.LoadLocation(userTimezone)
		if err == nil {
			loc = parsed
		}
	}

	now := time.Now().In(loc)
	today := s.truncateToDate(now, loc)
	yesterday := today.AddDate(0, 0, -1)

	var streak models.Streak
	result := s.db.Where("user_id = ?", userID).First(&streak)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to get streak: %w", result.Error)
	}

	if streak.LastSolvedDate == nil || streak.CurrentStreak == 0 {
		return &streak, true, nil // No active streak to check
	}

	lastSolved := s.truncateToDate(*streak.LastSolvedDate, loc)

	// Streak is valid if solved today or yesterday
	isValid := lastSolved.Equal(today) || lastSolved.Equal(yesterday)

	// If streak is broken, reset it
	if !isValid && streak.CurrentStreak > 0 {
		streak.CurrentStreak = 0
		if err := s.db.Save(&streak).Error; err != nil {
			return nil, false, fmt.Errorf("failed to reset streak: %w", err)
		}
	}

	return &streak, isValid, nil
}

// ResetStreak manually resets a user's current streak (e.g., for testing or admin purposes)
func (s *StreakService) ResetStreak(userID uint) error {
	result := s.db.Model(&models.Streak{}).
		Where("user_id = ?", userID).
		Update("current_streak", 0)

	if result.Error != nil {
		return fmt.Errorf("failed to reset streak: %w", result.Error)
	}
	return nil
}

// GetStreakStats returns aggregated streak statistics
func (s *StreakService) GetStreakStats() (*models.StreakStats, error) {
	stats := &models.StreakStats{
		StreakDistribution: make(map[string]int64),
	}

	// Total users with streak records
	if err := s.db.Model(&models.Streak{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Users with active streak (> 0)
	if err := s.db.Model(&models.Streak{}).
		Where("current_streak > 0").
		Count(&stats.UsersWithStreak).Error; err != nil {
		return nil, fmt.Errorf("failed to count active streaks: %w", err)
	}

	// Average streak (among those with streaks)
	var avgResult struct {
		Avg float64
	}
	if err := s.db.Model(&models.Streak{}).
		Select("COALESCE(AVG(current_streak), 0) as avg").
		Where("current_streak > 0").
		Scan(&avgResult).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average: %w", err)
	}
	stats.AverageStreak = avgResult.Avg

	// Max active streak
	var maxResult struct {
		Max int
	}
	if err := s.db.Model(&models.Streak{}).
		Select("COALESCE(MAX(current_streak), 0) as max").
		Scan(&maxResult).Error; err != nil {
		return nil, fmt.Errorf("failed to get max streak: %w", err)
	}
	stats.MaxActiveStreak = maxResult.Max

	// Streak distribution
	type distResult struct {
		Bucket string
		Count  int64
	}
	var distribution []distResult

	// Group streaks into buckets
	if err := s.db.Model(&models.Streak{}).
		Select(`CASE
			WHEN current_streak = 0 THEN '0'
			WHEN current_streak BETWEEN 1 AND 7 THEN '1-7'
			WHEN current_streak BETWEEN 8 AND 30 THEN '8-30'
			WHEN current_streak BETWEEN 31 AND 90 THEN '31-90'
			ELSE '90+'
		END as bucket, COUNT(*) as count`).
		Group("bucket").
		Scan(&distribution).Error; err != nil {
		return nil, fmt.Errorf("failed to get distribution: %w", err)
	}

	for _, d := range distribution {
		stats.StreakDistribution[d.Bucket] = d.Count
	}

	return stats, nil
}

// GetTopStreaks returns the top N users by current streak
func (s *StreakService) GetTopStreaks(limit int) ([]models.Streak, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var streaks []models.Streak
	err := s.db.Where("current_streak > 0").
		Order("current_streak DESC").
		Limit(limit).
		Find(&streaks).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get top streaks: %w", err)
	}

	return streaks, nil
}

// InitializeStreak creates an initial streak record for a new user
func (s *StreakService) InitializeStreak(userID uint) error {
	streak := models.Streak{
		UserID:        userID,
		CurrentStreak: 0,
		LongestStreak: 0,
	}

	// Use upsert to handle race conditions
	result := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoNothing: true,
	}).Create(&streak)

	if result.Error != nil {
		return fmt.Errorf("failed to initialize streak: %w", result.Error)
	}

	return nil
}

// truncateToDate truncates a time to the start of the day in the given timezone
func (s *StreakService) truncateToDate(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

// DaysUntilStreakBreaks calculates how many days until the streak would break
// Returns 0 if streak is already broken, 1 if must solve today, 2 if can skip today
func (s *StreakService) DaysUntilStreakBreaks(streak *models.Streak, userTimezone string) int {
	if streak == nil || streak.CurrentStreak == 0 || streak.LastSolvedDate == nil {
		return 0
	}

	loc := s.defaultTimezone
	if userTimezone != "" {
		parsed, err := time.LoadLocation(userTimezone)
		if err == nil {
			loc = parsed
		}
	}

	now := time.Now().In(loc)
	today := s.truncateToDate(now, loc)
	lastSolved := s.truncateToDate(*streak.LastSolvedDate, loc)

	// Calculate days since last solve
	daysSince := int(today.Sub(lastSolved).Hours() / 24)

	switch daysSince {
	case 0:
		// Solved today - safe until end of tomorrow
		return 2
	case 1:
		// Solved yesterday - must solve today
		return 1
	default:
		// Already broken
		return 0
	}
}
