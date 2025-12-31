package services

import (
	"fmt"
	"math"

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

// GetSubmissionByID retrieves a submission by its ID
func (s *SubmissionService) GetSubmissionByID(id uint) (*models.Submission, error) {
	var submission models.Submission

	result := s.db.First(&submission, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("submission with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to retrieve submission: %w", result.Error)
	}

	return &submission, nil
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

// CalculatePercentileMetrics computes percentile rankings for a submission
// Compares against all accepted submissions for the same problem
func (s *SubmissionService) CalculatePercentileMetrics(submissionID uint) (*models.PercentileMetrics, error) {
	// Get the target submission
	submission, err := s.GetSubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}

	// Only calculate percentiles for accepted submissions with metrics
	if submission.Status != models.StatusAccepted {
		return nil, fmt.Errorf("percentile metrics only available for accepted submissions")
	}

	metrics := &models.PercentileMetrics{
		SubmissionID: submission.ID,
		ProblemID:    submission.ProblemID,
	}

	// Get count of all accepted submissions for this problem
	var totalAccepted int64
	s.db.Model(&models.Submission{}).
		Where("problem_id = ? AND status = ?", submission.ProblemID, models.StatusAccepted).
		Count(&totalAccepted)

	metrics.TotalSubmissions = int(totalAccepted)

	// Calculate execution time percentile
	if submission.ExecutionTimeMs != nil {
		percentile, rank := s.calculateTimePercentile(submission.ProblemID, *submission.ExecutionTimeMs)
		metrics.ExecutionTimePercentile = percentile
		metrics.ExecutionTimeRank = rank

		if percentile != nil {
			metrics.ExecutionTimeMessage = formatPercentileMessage(*percentile, "faster")
		}
	}

	// Calculate memory usage percentile
	if submission.MemoryUsedKb != nil {
		percentile, rank := s.calculateMemoryPercentile(submission.ProblemID, *submission.MemoryUsedKb)
		metrics.MemoryPercentile = percentile
		metrics.MemoryRank = rank

		if percentile != nil {
			metrics.MemoryMessage = formatPercentileMessage(*percentile, "less memory")
		}
	}

	return metrics, nil
}

// calculateTimePercentile calculates what percentage of submissions are slower
// Returns (percentile, rank) where percentile is 0-100 and rank is 1-based position
func (s *SubmissionService) calculateTimePercentile(problemID uint, executionTimeMs int) (*float64, *int) {
	// Count submissions with slower or equal execution time
	var slowerCount int64
	s.db.Model(&models.Submission{}).
		Where("problem_id = ? AND status = ? AND execution_time_ms IS NOT NULL AND execution_time_ms >= ?",
			problemID, models.StatusAccepted, executionTimeMs).
		Count(&slowerCount)

	// Count submissions with strictly faster execution time (for rank)
	var fasterCount int64
	s.db.Model(&models.Submission{}).
		Where("problem_id = ? AND status = ? AND execution_time_ms IS NOT NULL AND execution_time_ms < ?",
			problemID, models.StatusAccepted, executionTimeMs).
		Count(&fasterCount)

	// Get total with valid execution time
	var total int64
	s.db.Model(&models.Submission{}).
		Where("problem_id = ? AND status = ? AND execution_time_ms IS NOT NULL",
			problemID, models.StatusAccepted).
		Count(&total)

	if total == 0 {
		return nil, nil
	}

	// Percentile = percentage of submissions that are slower
	// Exclude self from count for accurate comparison
	percentile := float64(slowerCount-1) / float64(total) * 100
	if percentile < 0 {
		percentile = 0
	}
	percentile = math.Round(percentile*100) / 100 // Round to 2 decimal places

	rank := int(fasterCount + 1)

	return &percentile, &rank
}

// calculateMemoryPercentile calculates what percentage of submissions use more memory
// Returns (percentile, rank) where percentile is 0-100 and rank is 1-based position
func (s *SubmissionService) calculateMemoryPercentile(problemID uint, memoryUsedKb int) (*float64, *int) {
	// Count submissions with higher or equal memory usage
	var higherCount int64
	s.db.Model(&models.Submission{}).
		Where("problem_id = ? AND status = ? AND memory_used_kb IS NOT NULL AND memory_used_kb >= ?",
			problemID, models.StatusAccepted, memoryUsedKb).
		Count(&higherCount)

	// Count submissions with strictly lower memory usage (for rank)
	var lowerCount int64
	s.db.Model(&models.Submission{}).
		Where("problem_id = ? AND status = ? AND memory_used_kb IS NOT NULL AND memory_used_kb < ?",
			problemID, models.StatusAccepted, memoryUsedKb).
		Count(&lowerCount)

	// Get total with valid memory metrics
	var total int64
	s.db.Model(&models.Submission{}).
		Where("problem_id = ? AND status = ? AND memory_used_kb IS NOT NULL",
			problemID, models.StatusAccepted).
		Count(&total)

	if total == 0 {
		return nil, nil
	}

	// Percentile = percentage of submissions that use more memory
	// Exclude self from count for accurate comparison
	percentile := float64(higherCount-1) / float64(total) * 100
	if percentile < 0 {
		percentile = 0
	}
	percentile = math.Round(percentile*100) / 100 // Round to 2 decimal places

	rank := int(lowerCount + 1)

	return &percentile, &rank
}

// formatPercentileMessage creates a human-readable percentile message
func formatPercentileMessage(percentile float64, comparison string) string {
	rounded := int(math.Round(percentile))
	if rounded >= 100 {
		rounded = 99
	}
	return fmt.Sprintf("%s than %d%% of submissions", comparison, rounded)
}

// GetProblemSubmissionStats returns aggregate statistics for a problem's submissions
func (s *SubmissionService) GetProblemSubmissionStats(problemID uint) (*ProblemSubmissionStats, error) {
	stats := &ProblemSubmissionStats{
		ProblemID: problemID,
	}

	// Count total submissions
	s.db.Model(&models.Submission{}).
		Where("problem_id = ?", problemID).
		Count(&stats.TotalSubmissions)

	// Count accepted submissions
	s.db.Model(&models.Submission{}).
		Where("problem_id = ? AND status = ?", problemID, models.StatusAccepted).
		Count(&stats.AcceptedSubmissions)

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
		Where("problem_id = ? AND status = ? AND execution_time_ms IS NOT NULL",
			problemID, models.StatusAccepted).
		Scan(&avgTime)
	stats.AvgExecutionTimeMs = avgTime.Avg

	// Get average memory usage for accepted submissions
	var avgMem struct {
		Avg *float64
	}
	s.db.Model(&models.Submission{}).
		Select("AVG(memory_used_kb) as avg").
		Where("problem_id = ? AND status = ? AND memory_used_kb IS NOT NULL",
			problemID, models.StatusAccepted).
		Scan(&avgMem)
	stats.AvgMemoryUsedKb = avgMem.Avg

	return stats, nil
}

// ProblemSubmissionStats contains aggregate statistics for a problem
type ProblemSubmissionStats struct {
	ProblemID           uint     `json:"problem_id"`
	TotalSubmissions    int64    `json:"total_submissions"`
	AcceptedSubmissions int64    `json:"accepted_submissions"`
	AcceptanceRate      float64  `json:"acceptance_rate"`
	AvgExecutionTimeMs  *float64 `json:"avg_execution_time_ms,omitempty"`
	AvgMemoryUsedKb     *float64 `json:"avg_memory_used_kb,omitempty"`
}
