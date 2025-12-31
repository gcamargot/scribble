package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/gorm"
)

// SubmissionService handles submission processing and streak integration
type SubmissionService struct {
	db               *gorm.DB
	streakService    *StreakService
	challengeService *DailyChallengeService
}

// NewSubmissionService creates a new submission service
func NewSubmissionService(db *gorm.DB, streakService *StreakService, challengeService *DailyChallengeService) *SubmissionService {
	return &SubmissionService{
		db:               db,
		streakService:    streakService,
		challengeService: challengeService,
	}
}

// ProcessSubmissionResult processes the result from the executor
// and triggers streak update if the submission is accepted for a daily challenge
func (s *SubmissionService) ProcessSubmissionResult(
	submissionID string,
	userID string,
	problemID uint,
	status string,
	compilationTimeMs int,
	executionTimeMs int,
	totalExecutionTimeMs int,
	memoryUsedKb int,
	testsPassed int,
	testsTotal int,
	errorMessage string,
	errorType string,
) (*models.Submission, *models.UserStreak, error) {
	// Update submission record
	submission := &models.Submission{
		ID:                   submissionID,
		UserID:               userID,
		ProblemID:            fmt.Sprintf("%d", problemID),
		Status:               status,
		CompilationTimeMs:    compilationTimeMs,
		ExecutionTimeMs:      executionTimeMs,
		TotalExecutionTimeMs: totalExecutionTimeMs,
		MemoryUsedKb:         memoryUsedKb,
		TestsPassed:          testsPassed,
		TestsTotal:           testsTotal,
		ErrorMessage:         errorMessage,
		ErrorType:            errorType,
		UpdatedAt:            time.Now(),
	}

	// Save or update submission
	if err := s.db.Save(submission).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to save submission: %w", err)
	}

	// Check if submission is accepted and trigger streak update
	var streak *models.UserStreak
	if status == models.StatusAccepted {
		// Try to update streak (will fail silently if not a daily challenge)
		updatedStreak, err := s.streakService.UpdateStreak(userID, problemID, submissionID)
		if err != nil {
			// Log but don't fail - streak update is not critical
			if !errors.Is(err, ErrNotDailyChallenge) && !errors.Is(err, ErrAlreadySolved) {
				fmt.Printf("Warning: failed to update streak: %v\n", err)
			}
		}
		streak = updatedStreak
	}

	return submission, streak, nil
}

// CreateSubmission creates a new pending submission
func (s *SubmissionService) CreateSubmission(userID, problemID, language, code string) (*models.Submission, error) {
	submission := &models.Submission{
		UserID:    userID,
		ProblemID: problemID,
		Language:  language,
		Code:      code,
		Status:    models.StatusPending,
	}

	if err := s.db.Create(submission).Error; err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	return submission, nil
}

// GetSubmission retrieves a submission by ID
func (s *SubmissionService) GetSubmission(id string) (*models.Submission, error) {
	var submission models.Submission
	if err := s.db.Where("id = ?", id).First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("submission not found")
		}
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}
	return &submission, nil
}

// GetUserSubmissions retrieves submissions for a user
func (s *SubmissionService) GetUserSubmissions(userID string, limit int) ([]models.Submission, error) {
	var submissions []models.Submission
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&submissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user submissions: %w", err)
	}
	return submissions, nil
}

// GetProblemSubmissions retrieves submissions for a problem
func (s *SubmissionService) GetProblemSubmissions(problemID string, limit int) ([]models.Submission, error) {
	var submissions []models.Submission
	err := s.db.Where("problem_id = ?", problemID).
		Order("created_at DESC").
		Limit(limit).
		Find(&submissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get problem submissions: %w", err)
	}
	return submissions, nil
}

// CheckDailyChallengeCompletion checks if a user has completed today's daily challenge
func (s *SubmissionService) CheckDailyChallengeCompletion(userID string) (bool, error) {
	// Get today's challenge
	challenge, err := s.challengeService.GetTodaysChallenge()
	if err != nil {
		return false, fmt.Errorf("failed to get today's challenge: %w", err)
	}
	if challenge == nil {
		return false, nil // No challenge today
	}

	// Check if user has an accepted submission for today's challenge
	var count int64
	err = s.db.Model(&models.Submission{}).
		Where("user_id = ? AND problem_id = ? AND status = ?",
			userID, fmt.Sprintf("%d", challenge.ProblemID), models.StatusAccepted).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check completion: %w", err)
	}

	return count > 0, nil
}
