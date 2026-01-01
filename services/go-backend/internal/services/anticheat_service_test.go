package services

import (
	"testing"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Migrate test tables
	err = db.AutoMigrate(&models.FlaggedSubmission{}, &models.RateLimitEntry{})
	if err != nil {
		t.Fatalf("failed to migrate test tables: %v", err)
	}

	return db
}

func TestIsSuspiciousTime(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	tests := []struct {
		name            string
		executionTimeMs int
		difficulty      string
		wantSuspicious  bool
	}{
		{
			name:            "easy problem with normal time",
			executionTimeMs: 100,
			difficulty:      "easy",
			wantSuspicious:  false,
		},
		{
			name:            "easy problem with suspicious time",
			executionTimeMs: 2,
			difficulty:      "easy",
			wantSuspicious:  true,
		},
		{
			name:            "easy problem at threshold",
			executionTimeMs: 5,
			difficulty:      "easy",
			wantSuspicious:  false,
		},
		{
			name:            "medium problem with normal time",
			executionTimeMs: 50,
			difficulty:      "medium",
			wantSuspicious:  false,
		},
		{
			name:            "medium problem with suspicious time",
			executionTimeMs: 5,
			difficulty:      "medium",
			wantSuspicious:  true,
		},
		{
			name:            "hard problem with normal time",
			executionTimeMs: 100,
			difficulty:      "hard",
			wantSuspicious:  false,
		},
		{
			name:            "hard problem with suspicious time",
			executionTimeMs: 10,
			difficulty:      "hard",
			wantSuspicious:  true,
		},
		{
			name:            "unknown difficulty uses default threshold",
			executionTimeMs: 3,
			difficulty:      "unknown",
			wantSuspicious:  true,
		},
		{
			name:            "zero execution time is always suspicious",
			executionTimeMs: 0,
			difficulty:      "easy",
			wantSuspicious:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isSuspiciousTime(tt.executionTimeMs, tt.difficulty)
			if got != tt.wantSuspicious {
				t.Errorf("isSuspiciousTime(%d, %s) = %v, want %v",
					tt.executionTimeMs, tt.difficulty, got, tt.wantSuspicious)
			}
		})
	}
}

func TestCheckSubmission_ZeroMemory(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	result, err := service.CheckSubmission(1, 1, 100, 0, "easy")
	if err != nil {
		t.Fatalf("CheckSubmission failed: %v", err)
	}

	if !result.Flagged {
		t.Error("expected submission to be flagged for zero memory")
	}

	foundZeroMemory := false
	for _, reason := range result.FlagReasons {
		if reason == models.FlagReasonZeroMemory {
			foundZeroMemory = true
			break
		}
	}
	if !foundZeroMemory {
		t.Error("expected FlagReasonZeroMemory in flag reasons")
	}
}

func TestCheckSubmission_SuspiciousTime(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	result, err := service.CheckSubmission(1, 1, 1, 1000, "hard")
	if err != nil {
		t.Fatalf("CheckSubmission failed: %v", err)
	}

	if !result.Flagged {
		t.Error("expected submission to be flagged for suspicious time")
	}

	foundSuspiciousTime := false
	for _, reason := range result.FlagReasons {
		if reason == models.FlagReasonSuspiciousTime {
			foundSuspiciousTime = true
			break
		}
	}
	if !foundSuspiciousTime {
		t.Error("expected FlagReasonSuspiciousTime in flag reasons")
	}
}

func TestCheckSubmission_NormalSubmission(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	result, err := service.CheckSubmission(1, 1, 100, 1000, "easy")
	if err != nil {
		t.Fatalf("CheckSubmission failed: %v", err)
	}

	if result.Flagged {
		t.Errorf("expected submission not to be flagged, got reasons: %v", result.FlagReasons)
	}

	if !result.Allowed {
		t.Error("expected submission to be allowed")
	}
}

func TestCheckSubmission_MultipleFlags(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	// Zero memory AND suspicious time
	result, err := service.CheckSubmission(1, 1, 1, 0, "hard")
	if err != nil {
		t.Fatalf("CheckSubmission failed: %v", err)
	}

	if !result.Flagged {
		t.Error("expected submission to be flagged")
	}

	if len(result.FlagReasons) != 2 {
		t.Errorf("expected 2 flag reasons, got %d: %v", len(result.FlagReasons), result.FlagReasons)
	}
}

func TestFlagSubmission(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	details := map[string]interface{}{
		"execution_time_ms": 1,
		"expected_min":      20,
	}

	err := service.FlagSubmission(1, 1, 1, models.FlagReasonSuspiciousTime, details)
	if err != nil {
		t.Fatalf("FlagSubmission failed: %v", err)
	}

	// Verify flag was created
	var flag models.FlaggedSubmission
	result := db.First(&flag, 1)
	if result.Error != nil {
		t.Fatalf("failed to retrieve flag: %v", result.Error)
	}

	if flag.SubmissionID != 1 {
		t.Errorf("expected SubmissionID 1, got %d", flag.SubmissionID)
	}
	if flag.Reason != models.FlagReasonSuspiciousTime {
		t.Errorf("expected reason %s, got %s", models.FlagReasonSuspiciousTime, flag.Reason)
	}
	if flag.Status != models.FlagStatusPending {
		t.Errorf("expected status %s, got %s", models.FlagStatusPending, flag.Status)
	}
}

func TestGetPendingFlags(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	// Create some flags
	flags := []models.FlaggedSubmission{
		{SubmissionID: 1, UserID: 1, ProblemID: 1, Reason: models.FlagReasonSuspiciousTime, Status: models.FlagStatusPending},
		{SubmissionID: 2, UserID: 1, ProblemID: 1, Reason: models.FlagReasonZeroMemory, Status: models.FlagStatusPending},
		{SubmissionID: 3, UserID: 2, ProblemID: 1, Reason: models.FlagReasonSuspiciousTime, Status: models.FlagStatusCleared},
	}
	for _, f := range flags {
		db.Create(&f)
	}

	// Test pagination
	result, total, err := service.GetPendingFlags(1, 10)
	if err != nil {
		t.Fatalf("GetPendingFlags failed: %v", err)
	}

	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestGetFlagsByUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	// Create flags for different users
	flags := []models.FlaggedSubmission{
		{SubmissionID: 1, UserID: 1, ProblemID: 1, Reason: models.FlagReasonSuspiciousTime, Status: models.FlagStatusPending},
		{SubmissionID: 2, UserID: 1, ProblemID: 2, Reason: models.FlagReasonZeroMemory, Status: models.FlagStatusCleared},
		{SubmissionID: 3, UserID: 2, ProblemID: 1, Reason: models.FlagReasonSuspiciousTime, Status: models.FlagStatusPending},
	}
	for _, f := range flags {
		db.Create(&f)
	}

	// Get user 1's flags
	result, err := service.GetFlagsByUser(1)
	if err != nil {
		t.Fatalf("GetFlagsByUser failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 flags for user 1, got %d", len(result))
	}

	// Get user 2's flags
	result, err = service.GetFlagsByUser(2)
	if err != nil {
		t.Fatalf("GetFlagsByUser failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 flag for user 2, got %d", len(result))
	}
}

func TestReviewFlag(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	// Create a flag
	flag := models.FlaggedSubmission{
		SubmissionID: 1,
		UserID:       1,
		ProblemID:    1,
		Reason:       models.FlagReasonSuspiciousTime,
		Status:       models.FlagStatusPending,
	}
	db.Create(&flag)

	// Review the flag
	err := service.ReviewFlag(flag.ID, 100, models.FlagStatusCleared, "False positive - legitimate solution")
	if err != nil {
		t.Fatalf("ReviewFlag failed: %v", err)
	}

	// Verify changes
	var updated models.FlaggedSubmission
	db.First(&updated, flag.ID)

	if updated.Status != models.FlagStatusCleared {
		t.Errorf("expected status %s, got %s", models.FlagStatusCleared, updated.Status)
	}

	if updated.ReviewedBy == nil || *updated.ReviewedBy != 100 {
		t.Error("expected ReviewedBy to be 100")
	}

	if updated.ReviewedAt == nil {
		t.Error("expected ReviewedAt to be set")
	}
}

func TestReviewFlag_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	err := service.ReviewFlag(999, 100, models.FlagStatusCleared, "test")
	if err == nil {
		t.Error("expected error for non-existent flag")
	}
}

func TestGetFlagStats(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	// Create flags with various statuses and reasons
	flags := []models.FlaggedSubmission{
		{SubmissionID: 1, UserID: 1, ProblemID: 1, Reason: models.FlagReasonSuspiciousTime, Status: models.FlagStatusPending},
		{SubmissionID: 2, UserID: 1, ProblemID: 1, Reason: models.FlagReasonSuspiciousTime, Status: models.FlagStatusPending},
		{SubmissionID: 3, UserID: 2, ProblemID: 1, Reason: models.FlagReasonZeroMemory, Status: models.FlagStatusCleared},
		{SubmissionID: 4, UserID: 3, ProblemID: 1, Reason: models.FlagReasonSuspiciousTime, Status: models.FlagStatusBanned},
	}
	for _, f := range flags {
		db.Create(&f)
	}

	stats, err := service.GetFlagStats()
	if err != nil {
		t.Fatalf("GetFlagStats failed: %v", err)
	}

	if stats.Total != 4 {
		t.Errorf("expected total 4, got %d", stats.Total)
	}

	if stats.Pending != 2 {
		t.Errorf("expected pending 2, got %d", stats.Pending)
	}

	if stats.Cleared != 1 {
		t.Errorf("expected cleared 1, got %d", stats.Cleared)
	}

	if stats.Banned != 1 {
		t.Errorf("expected banned 1, got %d", stats.Banned)
	}

	if stats.ByReason[models.FlagReasonSuspiciousTime] != 3 {
		t.Errorf("expected 3 suspicious_time flags, got %d", stats.ByReason[models.FlagReasonSuspiciousTime])
	}

	if stats.ByReason[models.FlagReasonZeroMemory] != 1 {
		t.Errorf("expected 1 zero_memory flag, got %d", stats.ByReason[models.FlagReasonZeroMemory])
	}
}

func TestCleanupOldRateLimitEntries(t *testing.T) {
	db := setupTestDB(t)
	service := NewAntiCheatService(db)

	// Create rate limit entries with different ages
	now := time.Now()
	oldTime := now.Add(-48 * time.Hour) // 2 days old
	recentTime := now.Add(-1 * time.Hour) // 1 hour old

	entries := []models.RateLimitEntry{
		{UserID: 1, Submissions: 5, WindowStart: oldTime, LastSubmit: oldTime},
		{UserID: 2, Submissions: 3, WindowStart: recentTime, LastSubmit: recentTime},
		{UserID: 3, Submissions: 2, WindowStart: oldTime, LastSubmit: oldTime},
	}
	for _, e := range entries {
		db.Create(&e)
	}

	// Cleanup
	deleted, err := service.CleanupOldRateLimitEntries()
	if err != nil {
		t.Fatalf("CleanupOldRateLimitEntries failed: %v", err)
	}

	if deleted != 2 {
		t.Errorf("expected 2 entries deleted, got %d", deleted)
	}

	// Verify only recent entry remains
	var remaining []models.RateLimitEntry
	db.Find(&remaining)

	if len(remaining) != 1 {
		t.Errorf("expected 1 remaining entry, got %d", len(remaining))
	}

	if remaining[0].UserID != 2 {
		t.Errorf("expected UserID 2 to remain, got %d", remaining[0].UserID)
	}
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := models.DefaultRateLimitConfig()

	if config.WindowDuration != 5*time.Minute {
		t.Errorf("expected WindowDuration 5 minutes, got %v", config.WindowDuration)
	}

	if config.MaxSubmissions != 10 {
		t.Errorf("expected MaxSubmissions 10, got %d", config.MaxSubmissions)
	}

	if config.CooldownDuration != 10*time.Minute {
		t.Errorf("expected CooldownDuration 10 minutes, got %v", config.CooldownDuration)
	}
}
