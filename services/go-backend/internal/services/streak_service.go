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
// Logic:
// - If user solves today's daily challenge (accepted), increment current_streak
// - If user missed a day (last_solved_date != yesterday), reset to 1
// - Update longest_streak if current exceeds it
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
		// Create new streak record
		streak = models.UserStreak{
			UserID: userID,
		}
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
		// First time solving
		streak.CurrentStreak = 1
	} else {
		lastSolved := streak.LastSolvedDate.Truncate(24 * time.Hour)
		if lastSolved.Equal(yesterday) {
			// Consecutive day - extend streak
			streak.CurrentStreak++
		} else {
			// Missed a day - reset streak
			streak.CurrentStreak = 1
		}
	}

	// Update longest streak if exceeded
	if streak.CurrentStreak > streak.LongestStreak {
		streak.LongestStreak = streak.CurrentStreak
	}

	// Update last solved date and total days
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
		// Log but don't fail - history is for analytics
		fmt.Printf("Warning: failed to record streak history: %v\n", err)
	}

	return &streak, nil
}

// GetStreak returns a user's current streak information
func (s *StreakService) GetStreak(userID string) (*models.UserStreak, error) {
	var streak models.UserStreak
	err := s.db.Where("user_id = ?", userID).First(&streak).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Return empty streak if not found
		return &models.UserStreak{
			UserID:        userID,
			CurrentStreak: 0,
			LongestStreak: 0,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get streak: %w", err)
	}

	// Check if streak is still valid (solved yesterday or today)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	if streak.LastSolvedDate != nil {
		lastSolved := streak.LastSolvedDate.Truncate(24 * time.Hour)
		if !lastSolved.Equal(today) && !lastSolved.Equal(yesterday) {
			// Streak has expired - reset current but keep longest
			streak.CurrentStreak = 0
		}
	}

	return &streak, nil
}

// GetLeaderboard returns top users by streak (current or longest)
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

// GetStreakHistory returns a user's streak history
func (s *StreakService) GetStreakHistory(userID string, limit int) ([]models.StreakHistory, error) {
	var history []models.StreakHistory
	err := s.db.Where("user_id = ?", userID).
		Order("solved_date DESC").
		Limit(limit).
		Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get streak history: %w", err)
	}
	return history, nil
}
