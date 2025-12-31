package services

import (
	"fmt"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/gorm"
)

// ProblemService handles business logic for problem operations
type ProblemService struct {
	db *gorm.DB
}

// NewProblemService creates a new problem service instance
func NewProblemService(db *gorm.DB) *ProblemService {
	return &ProblemService{
		db: db,
	}
}

// GetProblemByID retrieves a problem by its ID
// Returns error if problem not found
func (s *ProblemService) GetProblemByID(id uint) (*models.Problem, error) {
	var problem models.Problem

	// Query problem by ID
	result := s.db.First(&problem, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("problem with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve problem: %w", result.Error)
	}

	return &problem, nil
}

// GetTestCasesByProblemID retrieves test cases for a specific problem
// If sampleOnly is true, returns only sample test cases (hides hidden tests)
func (s *ProblemService) GetTestCasesByProblemID(problemID uint, sampleOnly bool) ([]models.TestCase, error) {
	var testCases []models.TestCase

	query := s.db.Where("problem_id = ?", problemID)

	// Filter to sample tests only if requested (for user-facing endpoints)
	if sampleOnly {
		query = query.Where("is_sample = ?", true)
	}

	// Execute query ordered by ID for consistent ordering
	result := query.Order("id ASC").Find(&testCases)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve test cases: %w", result.Error)
	}

	return testCases, nil
}

// GetDailyChallengeByDate retrieves the daily challenge for a specific date
// Date should be in YYYY-MM-DD format (UTC)
// Returns the challenge with the associated problem preloaded
func (s *ProblemService) GetDailyChallengeByDate(date time.Time) (*models.DailyChallenge, error) {
	var challenge models.DailyChallenge

	// Truncate time to date only (midnight UTC)
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	// Query daily challenge by date and preload the associated problem
	result := s.db.Preload("Problem").Where("challenge_date = ?", dateOnly).First(&challenge)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no daily challenge found for date %s", dateOnly.Format("2006-01-02"))
		}
		return nil, fmt.Errorf("failed to retrieve daily challenge: %w", result.Error)
	}

	return &challenge, nil
}

// GetTodaysDailyChallenge is a convenience method to get today's challenge
// Uses UTC timezone for consistency
func (s *ProblemService) GetTodaysDailyChallenge() (*models.DailyChallenge, error) {
	today := time.Now().UTC()
	return s.GetDailyChallengeByDate(today)
}
