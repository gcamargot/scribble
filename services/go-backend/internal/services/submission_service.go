package services

import (
	"fmt"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/gorm"
)

// SubmissionService handles business logic for submission operations
type SubmissionService struct {
	db *gorm.DB
}

// NewSubmissionService creates a new submission service instance
func NewSubmissionService(db *gorm.DB) *SubmissionService {
	return &SubmissionService{
		db: db,
	}
}

// CreateSubmission saves a new submission to the database
func (s *SubmissionService) CreateSubmission(submission *models.Submission) error {
	result := s.db.Create(submission)
	if result.Error != nil {
		return fmt.Errorf("failed to create submission: %w", result.Error)
	}
	return nil
}

// GetSubmissionByID retrieves a submission by its ID
func (s *SubmissionService) GetSubmissionByID(id uint) (*models.Submission, error) {
	var submission models.Submission

	result := s.db.Preload("Problem").First(&submission, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("submission with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve submission: %w", result.Error)
	}

	return &submission, nil
}

// GetSubmissionWithCode retrieves a submission with its code included
func (s *SubmissionService) GetSubmissionWithCode(id uint) (*models.SubmissionWithCode, error) {
	var submission models.Submission

	result := s.db.Preload("Problem").First(&submission, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("submission with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve submission: %w", result.Error)
	}

	// We need to get the code separately since it's not in the default response
	var codeResult struct {
		Code string
	}
	s.db.Model(&models.Submission{}).Select("code").Where("id = ?", id).Scan(&codeResult)

	return submission.ToWithCode(codeResult.Code), nil
}

// SubmissionHistoryParams contains pagination and filter parameters
type SubmissionHistoryParams struct {
	UserID    uint
	Page      int    // 1-indexed
	PageSize  int    // Default 20
	ProblemID *uint  // Optional filter by problem
	Status    string // Optional filter by status
	Language  string // Optional filter by language
}

// SubmissionHistoryResult contains paginated submission history
type SubmissionHistoryResult struct {
	Submissions []models.Submission `json:"submissions"`
	Total       int64               `json:"total"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
	TotalPages  int                 `json:"total_pages"`
}

// GetUserSubmissionHistory retrieves paginated submission history for a user
func (s *SubmissionService) GetUserSubmissionHistory(params SubmissionHistoryParams) (*SubmissionHistoryResult, error) {
	// Apply defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	// Build query
	query := s.db.Model(&models.Submission{}).Where("user_id = ?", params.UserID)

	// Apply optional filters
	if params.ProblemID != nil {
		query = query.Where("problem_id = ?", *params.ProblemID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Language != "" {
		query = query.Where("language = ?", params.Language)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count submissions: %w", err)
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	// Fetch submissions with problem info
	var submissions []models.Submission
	result := query.
		Preload("Problem").
		Order("submitted_at DESC").
		Offset(offset).
		Limit(params.PageSize).
		Find(&submissions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve submissions: %w", result.Error)
	}

	return &SubmissionHistoryResult{
		Submissions: submissions,
		Total:       total,
		Page:        params.Page,
		PageSize:    params.PageSize,
		TotalPages:  totalPages,
	}, nil
}

// GetSubmissionsByUserAndProblem retrieves all submissions by a user for a specific problem
func (s *SubmissionService) GetSubmissionsByUserAndProblem(userID, problemID uint) ([]models.Submission, error) {
	var submissions []models.Submission

	result := s.db.Where("user_id = ? AND problem_id = ?", userID, problemID).
		Order("submitted_at DESC").
		Find(&submissions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve submissions: %w", result.Error)
	}

	return submissions, nil
}

// GetUserSubmissionStats returns submission statistics for a user
func (s *SubmissionService) GetUserSubmissionStats(userID uint) (*UserSubmissionStats, error) {
	stats := &UserSubmissionStats{
		UserID: userID,
	}

	// Count total submissions
	s.db.Model(&models.Submission{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalSubmissions)

	// Count accepted submissions
	s.db.Model(&models.Submission{}).
		Where("user_id = ? AND status = ?", userID, models.StatusAccepted).
		Count(&stats.AcceptedSubmissions)

	// Count unique problems solved (accepted at least once)
	s.db.Model(&models.Submission{}).
		Select("COUNT(DISTINCT problem_id)").
		Where("user_id = ? AND status = ?", userID, models.StatusAccepted).
		Scan(&stats.ProblemsSolved)

	// Calculate acceptance rate
	if stats.TotalSubmissions > 0 {
		stats.AcceptanceRate = float64(stats.AcceptedSubmissions) / float64(stats.TotalSubmissions) * 100
	}

	// Get average execution time for accepted submissions
	var avgTime struct {
		Avg *float64
	}
	s.db.Model(&models.Submission{}).
		Select("AVG(execution_time_ms) as avg").
		Where("user_id = ? AND status = ? AND execution_time_ms IS NOT NULL",
			userID, models.StatusAccepted).
		Scan(&avgTime)
	stats.AvgExecutionTimeMs = avgTime.Avg

	// Get average memory usage for accepted submissions
	var avgMem struct {
		Avg *float64
	}
	s.db.Model(&models.Submission{}).
		Select("AVG(memory_used_kb) as avg").
		Where("user_id = ? AND status = ? AND memory_used_kb IS NOT NULL",
			userID, models.StatusAccepted).
		Scan(&avgMem)
	stats.AvgMemoryUsedKb = avgMem.Avg

	return stats, nil
}

// UserSubmissionStats contains aggregate statistics for a user
type UserSubmissionStats struct {
	UserID              uint     `json:"user_id"`
	TotalSubmissions    int64    `json:"total_submissions"`
	AcceptedSubmissions int64    `json:"accepted_submissions"`
	ProblemsSolved      int64    `json:"problems_solved"`
	AcceptanceRate      float64  `json:"acceptance_rate"`
	AvgExecutionTimeMs  *float64 `json:"avg_execution_time_ms,omitempty"`
	AvgMemoryUsedKb     *float64 `json:"avg_memory_used_kb,omitempty"`
}
