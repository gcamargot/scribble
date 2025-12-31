package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/gorm"
)

// Common errors for daily challenge service
var (
	ErrChallengeExists = errors.New("daily challenge already exists for today")
	ErrNoProblems      = errors.New("no problems available in database")
)

// DailyChallengeService handles daily challenge selection and management
type DailyChallengeService struct {
	db *gorm.DB
}

// NewDailyChallengeService creates a new daily challenge service
func NewDailyChallengeService(db *gorm.DB) *DailyChallengeService {
	return &DailyChallengeService{db: db}
}

// SelectNextChallenge selects the next problem for daily challenge
func (s *DailyChallengeService) SelectNextChallenge() (*models.DailyChallenge, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	var existing models.DailyChallenge
	err := s.db.Where("challenge_date = ?", today).First(&existing).Error
	if err == nil {
		return &existing, ErrChallengeExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing challenge: %w", err)
	}

	var problemCount int64
	if err := s.db.Model(&models.Problem{}).Count(&problemCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count problems: %w", err)
	}
	if problemCount == 0 {
		return nil, ErrNoProblems
	}

	var nextProblem models.Problem
	subQuery := s.db.Model(&models.DailyChallenge{}).Select("problem_id")
	err = s.db.Where("id NOT IN (?)", subQuery).Order("id ASC").First(&nextProblem).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = s.db.
			Joins("LEFT JOIN daily_challenges ON problems.id = daily_challenges.problem_id").
			Group("problems.id").
			Order("MAX(daily_challenges.challenge_date) ASC NULLS FIRST, problems.id ASC").
			First(&nextProblem).Error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to select next problem: %w", err)
	}

	challenge := &models.DailyChallenge{
		ProblemID:     nextProblem.ID,
		ChallengeDate: today,
	}

	if err := s.db.Create(challenge).Error; err != nil {
		return nil, fmt.Errorf("failed to create daily challenge: %w", err)
	}

	challenge.Problem = nextProblem
	return challenge, nil
}

// GetTodaysChallenge returns today's daily challenge
func (s *DailyChallengeService) GetTodaysChallenge() (*models.DailyChallenge, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	var challenge models.DailyChallenge
	err := s.db.Preload("Problem").Where("challenge_date = ?", today).First(&challenge).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get today's challenge: %w", err)
	}

	return &challenge, nil
}

// GetChallengeByDate returns the daily challenge for a specific date
func (s *DailyChallengeService) GetChallengeByDate(date time.Time) (*models.DailyChallenge, error) {
	dateOnly := date.UTC().Truncate(24 * time.Hour)

	var challenge models.DailyChallenge
	err := s.db.Preload("Problem").Where("challenge_date = ?", dateOnly).First(&challenge).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get challenge for date: %w", err)
	}

	return &challenge, nil
}
