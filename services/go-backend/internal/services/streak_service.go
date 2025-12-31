package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/gorm"
)

// Common errors for streak service
var (
	ErrNotDailyChallenge = errors.New("submission is not for today's daily challenge")
	ErrAlreadySolved     = errors.New("user already solved today's daily challenge")
)

// StreakService handles user streak management
type StreakService struct {
	db               *gorm.DB
	challengeService *DailyChallengeService
}

// NewStreakService creates a new streak service
func NewStreakService(db *gorm.DB, challengeService *DailyChallengeService) *StreakService {
	return &StreakService{
		db:               db,
		challengeService: challengeService,
	}
}

// UpdateStreak updates a user's streak after solving the daily challenge
func (s *StreakService) UpdateStreak(userID string, problemID uint, submissionID string) (*models.UserStreak, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	// Check if this is today's daily challenge
	todaysChallenge, err := s.challengeService.GetTodaysChallenge()
	if err != nil {
		return nil, fmt.Errorf("failed to get today's challenge: %w", err)
	}
	if todaysChallenge == nil {
		return nil, ErrNotDailyChallenge
	}
	if todaysChallenge.ProblemID != problemID {
		return nil, ErrNotDailyChallenge
	}

	// Get or create user streak record
	var streak models.UserStreak
	err = s.db.Where("user_id = ?", userID).First(&streak).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		streak = models.UserStreak{UserID: userID}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user streak: %w", err)
	}

	// Check if already solved today
	if streak.LastSolvedDate != nil {
		lastSolved := streak.LastSolvedDate.Truncate(24 * time.Hour)
		if lastSolved.Equal(today) {
			return &streak, ErrAlreadySolved
		}
	}

	// Update streak based on last solved date
	if streak.LastSolvedDate == nil {
		streak.CurrentStreak = 1
	} else {
		lastSolved := streak.LastSolvedDate.Truncate(24 * time.Hour)
		if lastSolved.Equal(yesterday) {
			streak.CurrentStreak++
		} else {
			streak.CurrentStreak = 1
		}
	}

	// Update longest streak if exceeded
	if streak.CurrentStreak > streak.LongestStreak {
		streak.LongestStreak = streak.CurrentStreak
	}

	streak.LastSolvedDate = &today
	streak.TotalDaysSolved++

	// Save streak
	if streak.ID == 0 {
		err = s.db.Create(&streak).Error
	} else {
		err = s.db.Save(&streak).Error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update streak: %w", err)
	}

	// Record in streak history
	history := models.StreakHistory{
		UserID:       userID,
		SolvedDate:   today,
		ProblemID:    problemID,
		SubmissionID: submissionID,
		StreakDay:    streak.CurrentStreak,
	}
	if err := s.db.Create(&history).Error; err != nil {
		fmt.Printf("Warning: failed to record streak history: %v\n", err)
	}

	return &streak, nil
}

// GetStreak returns a user's current streak information
func (s *StreakService) GetStreak(userID string) (*models.UserStreak, error) {
	var streak models.UserStreak
	err := s.db.Where("user_id = ?", userID).First(&streak).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.UserStreak{
			UserID:        userID,
			CurrentStreak: 0,
			LongestStreak: 0,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get streak: %w", err)
	}

	// Check if streak is still valid
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	if streak.LastSolvedDate != nil {
		lastSolved := streak.LastSolvedDate.Truncate(24 * time.Hour)
		if !lastSolved.Equal(today) && !lastSolved.Equal(yesterday) {
			streak.CurrentStreak = 0
		}
	}

	return &streak, nil
}

// GetLeaderboard returns top users by streak
func (s *StreakService) GetLeaderboard(limit int, byLongest bool) ([]models.UserStreak, error) {
	var streaks []models.UserStreak

	orderBy := "current_streak DESC"
	if byLongest {
		orderBy = "longest_streak DESC"
	}

	err := s.db.Order(orderBy).Limit(limit).Find(&streaks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return streaks, nil
}
