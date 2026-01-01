package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/gorm"
)

// AntiCheatService handles cheating detection and prevention
type AntiCheatService struct {
	db              *gorm.DB
	rateLimitConfig models.RateLimitConfig
}

// NewAntiCheatService creates a new anti-cheat service instance
func NewAntiCheatService(db *gorm.DB) *AntiCheatService {
	return &AntiCheatService{
		db:              db,
		rateLimitConfig: models.DefaultRateLimitConfig(),
	}
}

// SubmissionCheckResult contains the result of checking a submission
type SubmissionCheckResult struct {
	Allowed      bool              `json:"allowed"`
	Flagged      bool              `json:"flagged"`
	FlagReasons  []models.FlagReason `json:"flag_reasons,omitempty"`
	RateLimited  bool              `json:"rate_limited"`
	Message      string            `json:"message,omitempty"`
	RetryAfter   *time.Duration    `json:"retry_after,omitempty"`
}

// CheckSubmission performs anti-cheat checks on a submission
// Called before or after code execution
func (s *AntiCheatService) CheckSubmission(userID, problemID uint, executionTimeMs, memoryUsedKb int, difficulty string) (*SubmissionCheckResult, error) {
	result := &SubmissionCheckResult{
		Allowed:     true,
		Flagged:     false,
		FlagReasons: []models.FlagReason{},
	}

	// Check rate limit first
	rateLimited, retryAfter, err := s.checkRateLimit(userID)
	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	if rateLimited {
		result.Allowed = false
		result.RateLimited = true
		result.RetryAfter = retryAfter
		result.Message = "Rate limit exceeded. Please wait before submitting again."
		return result, nil
	}

	// Check for suspicious execution time
	if s.isSuspiciousTime(executionTimeMs, difficulty) {
		result.Flagged = true
		result.FlagReasons = append(result.FlagReasons, models.FlagReasonSuspiciousTime)
	}

	// Check for zero memory (impossible in real execution)
	if memoryUsedKb == 0 {
		result.Flagged = true
		result.FlagReasons = append(result.FlagReasons, models.FlagReasonZeroMemory)
	}

	return result, nil
}

// isSuspiciousTime checks if execution time is too fast for problem difficulty
func (s *AntiCheatService) isSuspiciousTime(executionTimeMs int, difficulty string) bool {
	threshold, ok := models.SuspiciousTimeThresholds[difficulty]
	if !ok {
		threshold = 5 // Default to 5ms for unknown difficulty
	}

	return executionTimeMs < threshold
}

// checkRateLimit checks if user has exceeded submission rate limit
// Uses atomic updates to prevent race conditions
func (s *AntiCheatService) checkRateLimit(userID uint) (bool, *time.Duration, error) {
	now := time.Now()
	windowStart := now.Add(-s.rateLimitConfig.WindowDuration)

	// Try atomic increment for existing entry within window
	// This prevents race conditions by using UPDATE ... WHERE in a single query
	result := s.db.Exec(`
		UPDATE rate_limit_entries
		SET submissions = submissions + 1, last_submit = ?
		WHERE user_id = ?
		  AND window_start > ?
		  AND submissions < ?
	`, now, userID, windowStart, s.rateLimitConfig.MaxSubmissions)

	if result.Error != nil {
		return false, nil, fmt.Errorf("failed to update rate limit: %w", result.Error)
	}

	// If we updated a row, the submission is allowed
	if result.RowsAffected > 0 {
		return false, nil, nil
	}

	// No row updated - either entry doesn't exist, window expired, or limit reached
	// Need to check which case and handle accordingly
	var entry models.RateLimitEntry
	err := s.db.Where("user_id = ?", userID).First(&entry).Error

	if err == gorm.ErrRecordNotFound {
		// First submission - create entry atomically using INSERT ... ON CONFLICT
		// Use ON CONFLICT to handle race between concurrent first submissions
		createResult := s.db.Exec(`
			INSERT INTO rate_limit_entries (user_id, submissions, window_start, last_submit)
			VALUES (?, 1, ?, ?)
			ON CONFLICT (user_id) DO UPDATE SET
				submissions = CASE
					WHEN rate_limit_entries.window_start <= ? THEN 1
					WHEN rate_limit_entries.submissions < ? THEN rate_limit_entries.submissions + 1
					ELSE rate_limit_entries.submissions
				END,
				window_start = CASE
					WHEN rate_limit_entries.window_start <= ? THEN ?
					ELSE rate_limit_entries.window_start
				END,
				last_submit = ?
		`, userID, now, now, windowStart, s.rateLimitConfig.MaxSubmissions, windowStart, now, now)

		if createResult.Error != nil {
			return false, nil, fmt.Errorf("failed to create rate limit entry: %w", createResult.Error)
		}
		return false, nil, nil
	} else if err != nil {
		return false, nil, fmt.Errorf("failed to query rate limit: %w", err)
	}

	// Entry exists - check if window expired
	windowEnd := entry.WindowStart.Add(s.rateLimitConfig.WindowDuration)
	if now.After(windowEnd) {
		// Window expired - reset atomically
		s.db.Exec(`
			UPDATE rate_limit_entries
			SET submissions = 1, window_start = ?, last_submit = ?
			WHERE user_id = ? AND window_start = ?
		`, now, now, userID, entry.WindowStart)
		return false, nil, nil
	}

	// Limit reached - check cooldown
	if entry.Submissions >= s.rateLimitConfig.MaxSubmissions {
		cooldownEnd := entry.LastSubmit.Add(s.rateLimitConfig.CooldownDuration)
		if now.Before(cooldownEnd) {
			remaining := cooldownEnd.Sub(now)
			return true, &remaining, nil
		}
		// Cooldown expired - reset atomically
		s.db.Exec(`
			UPDATE rate_limit_entries
			SET submissions = 1, window_start = ?, last_submit = ?
			WHERE user_id = ? AND submissions >= ?
		`, now, now, userID, s.rateLimitConfig.MaxSubmissions)
		return false, nil, nil
	}

	// Edge case: concurrent request incremented just before us
	return false, nil, nil
}

// FlagSubmission creates a flag record for a suspicious submission
func (s *AntiCheatService) FlagSubmission(submissionID, userID, problemID uint, reason models.FlagReason, details map[string]interface{}) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	flag := models.FlaggedSubmission{
		SubmissionID: submissionID,
		UserID:       userID,
		ProblemID:    problemID,
		Reason:       reason,
		Details:      string(detailsJSON),
		Status:       models.FlagStatusPending,
	}

	result := s.db.Create(&flag)
	if result.Error != nil {
		return fmt.Errorf("failed to create flag: %w", result.Error)
	}

	return nil
}

// GetPendingFlags retrieves flagged submissions awaiting review
func (s *AntiCheatService) GetPendingFlags(page, pageSize int) ([]models.FlaggedSubmission, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	s.db.Model(&models.FlaggedSubmission{}).Where("status = ?", models.FlagStatusPending).Count(&total)

	var flags []models.FlaggedSubmission
	offset := (page - 1) * pageSize

	result := s.db.Where("status = ?", models.FlagStatusPending).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&flags)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to get pending flags: %w", result.Error)
	}

	return flags, total, nil
}

// GetFlagsByUser retrieves all flags for a specific user
func (s *AntiCheatService) GetFlagsByUser(userID uint) ([]models.FlaggedSubmission, error) {
	var flags []models.FlaggedSubmission

	result := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&flags)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user flags: %w", result.Error)
	}

	return flags, nil
}

// ReviewFlag updates the status of a flagged submission (admin action)
func (s *AntiCheatService) ReviewFlag(flagID, adminUserID uint, status models.FlagStatus, notes string) error {
	now := time.Now()

	result := s.db.Model(&models.FlaggedSubmission{}).
		Where("id = ?", flagID).
		Updates(map[string]interface{}{
			"status":       status,
			"reviewed_by":  adminUserID,
			"reviewed_at":  now,
			"review_notes": notes,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to review flag: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("flag with ID %d not found", flagID)
	}

	return nil
}

// GetFlagStats returns statistics about flagged submissions
func (s *AntiCheatService) GetFlagStats() (*FlagStats, error) {
	stats := &FlagStats{}

	// Count by status
	s.db.Model(&models.FlaggedSubmission{}).Where("status = ?", models.FlagStatusPending).Count(&stats.Pending)
	s.db.Model(&models.FlaggedSubmission{}).Where("status = ?", models.FlagStatusReviewed).Count(&stats.Reviewed)
	s.db.Model(&models.FlaggedSubmission{}).Where("status = ?", models.FlagStatusCleared).Count(&stats.Cleared)
	s.db.Model(&models.FlaggedSubmission{}).Where("status = ?", models.FlagStatusBanned).Count(&stats.Banned)
	s.db.Model(&models.FlaggedSubmission{}).Count(&stats.Total)

	// Count by reason
	type reasonCount struct {
		Reason models.FlagReason
		Count  int64
	}
	var reasonCounts []reasonCount

	s.db.Model(&models.FlaggedSubmission{}).
		Select("reason, COUNT(*) as count").
		Group("reason").
		Scan(&reasonCounts)

	stats.ByReason = make(map[models.FlagReason]int64)
	for _, rc := range reasonCounts {
		stats.ByReason[rc.Reason] = rc.Count
	}

	return stats, nil
}

// FlagStats contains aggregated flag statistics
type FlagStats struct {
	Total    int64                        `json:"total"`
	Pending  int64                        `json:"pending"`
	Reviewed int64                        `json:"reviewed"`
	Cleared  int64                        `json:"cleared"`
	Banned   int64                        `json:"banned"`
	ByReason map[models.FlagReason]int64  `json:"by_reason"`
}

// CleanupOldRateLimitEntries removes stale rate limit entries (called periodically)
func (s *AntiCheatService) CleanupOldRateLimitEntries() (int64, error) {
	cutoff := time.Now().Add(-24 * time.Hour) // Remove entries older than 24 hours

	result := s.db.Where("last_submit < ?", cutoff).Delete(&models.RateLimitEntry{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup rate limit entries: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// EnsureTables creates the required tables if they don't exist
func (s *AntiCheatService) EnsureTables() error {
	return s.db.AutoMigrate(&models.FlaggedSubmission{}, &models.RateLimitEntry{})
}
