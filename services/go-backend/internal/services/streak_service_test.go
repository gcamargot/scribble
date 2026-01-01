package services

import (
	"testing"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupStreakTestDB creates an in-memory SQLite database for testing
func setupStreakTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Streak{})
	if err != nil {
		t.Fatalf("failed to migrate test tables: %v", err)
	}

	return db
}

func TestNewStreakService(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	if service == nil {
		t.Fatal("expected service to be created")
	}
	if service.defaultTimezone != time.UTC {
		t.Error("expected default timezone to be UTC")
	}
}

func TestNewStreakServiceWithTimezone(t *testing.T) {
	db := setupStreakTestDB(t)
	loc, _ := time.LoadLocation("America/New_York")
	service := NewStreakServiceWithTimezone(db, loc)

	if service.defaultTimezone != loc {
		t.Error("expected custom timezone to be set")
	}
}

func TestUpdateStreak_FirstSolve(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	update, err := service.UpdateStreak(1, "")
	if err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}

	if update.CurrentStreak != 1 {
		t.Errorf("expected current streak 1, got %d", update.CurrentStreak)
	}
	if update.PreviousStreak != 0 {
		t.Errorf("expected previous streak 0, got %d", update.PreviousStreak)
	}
	if update.WasReset {
		t.Error("expected WasReset to be false for first solve")
	}

	// Verify in database
	streak, err := service.GetUserStreak(1)
	if err != nil {
		t.Fatalf("GetUserStreak failed: %v", err)
	}
	if streak.CurrentStreak != 1 {
		t.Errorf("expected streak 1 in DB, got %d", streak.CurrentStreak)
	}
}

func TestUpdateStreak_SameDay(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// First solve
	update1, _ := service.UpdateStreak(1, "")

	// Solve again same day
	update2, err := service.UpdateStreak(1, "")
	if err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}

	// Streak should not change
	if update2.CurrentStreak != update1.CurrentStreak {
		t.Errorf("expected streak unchanged, got %d", update2.CurrentStreak)
	}
}

func TestUpdateStreak_ConsecutiveDays(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Create streak with last solved yesterday
	loc := time.UTC
	yesterday := time.Now().In(loc).AddDate(0, 0, -1)
	streak := models.Streak{
		UserID:         1,
		CurrentStreak:  5,
		LongestStreak:  5,
		LastSolvedDate: &yesterday,
	}
	db.Create(&streak)

	// Solve today
	update, err := service.UpdateStreak(1, "")
	if err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}

	if update.CurrentStreak != 6 {
		t.Errorf("expected streak to continue to 6, got %d", update.CurrentStreak)
	}
	if update.WasReset {
		t.Error("expected WasReset to be false for consecutive day")
	}
}

func TestUpdateStreak_BrokenStreak(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Create streak with last solved 3 days ago
	loc := time.UTC
	threeDaysAgo := time.Now().In(loc).AddDate(0, 0, -3)
	streak := models.Streak{
		UserID:         1,
		CurrentStreak:  10,
		LongestStreak:  10,
		LastSolvedDate: &threeDaysAgo,
	}
	db.Create(&streak)

	// Solve today
	update, err := service.UpdateStreak(1, "")
	if err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}

	if update.CurrentStreak != 1 {
		t.Errorf("expected streak to reset to 1, got %d", update.CurrentStreak)
	}
	if !update.WasReset {
		t.Error("expected WasReset to be true for broken streak")
	}
	if update.PreviousStreak != 10 {
		t.Errorf("expected previous streak 10, got %d", update.PreviousStreak)
	}
}

func TestUpdateStreak_NewRecord(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Create streak about to break record
	loc := time.UTC
	yesterday := time.Now().In(loc).AddDate(0, 0, -1)
	streak := models.Streak{
		UserID:         1,
		CurrentStreak:  10,
		LongestStreak:  10,
		LastSolvedDate: &yesterday,
	}
	db.Create(&streak)

	update, err := service.UpdateStreak(1, "")
	if err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}

	if !update.IsNewRecord {
		t.Error("expected IsNewRecord to be true when breaking record")
	}
	if update.LongestStreak != 11 {
		t.Errorf("expected longest streak 11, got %d", update.LongestStreak)
	}
}

func TestUpdateStreak_Timezone_NewYork(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Solve using New York timezone
	update, err := service.UpdateStreak(1, "America/New_York")
	if err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}

	if update.CurrentStreak != 1 {
		t.Errorf("expected streak 1, got %d", update.CurrentStreak)
	}

	// The solve date should be normalized to the user's timezone
	nyLoc, _ := time.LoadLocation("America/New_York")
	expectedDate := service.truncateToDate(time.Now().In(nyLoc), nyLoc)
	if !update.SolvedDate.Equal(expectedDate) {
		t.Errorf("expected solved date %v, got %v", expectedDate, update.SolvedDate)
	}
}

func TestUpdateStreak_Timezone_Tokyo(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	update, err := service.UpdateStreak(1, "Asia/Tokyo")
	if err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}

	if update.CurrentStreak != 1 {
		t.Errorf("expected streak 1, got %d", update.CurrentStreak)
	}

	tokyoLoc, _ := time.LoadLocation("Asia/Tokyo")
	expectedDate := service.truncateToDate(time.Now().In(tokyoLoc), tokyoLoc)
	if !update.SolvedDate.Equal(expectedDate) {
		t.Errorf("expected solved date in Tokyo time %v, got %v", expectedDate, update.SolvedDate)
	}
}

func TestUpdateStreak_InvalidTimezone(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Invalid timezone should fall back to UTC
	update, err := service.UpdateStreak(1, "Invalid/Timezone")
	if err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}

	// Should still work with default UTC timezone
	if update.CurrentStreak != 1 {
		t.Errorf("expected streak 1, got %d", update.CurrentStreak)
	}
}

func TestCheckStreak_Valid(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Create streak solved today
	loc := time.UTC
	today := service.truncateToDate(time.Now().In(loc), loc)
	streak := models.Streak{
		UserID:         1,
		CurrentStreak:  5,
		LongestStreak:  5,
		LastSolvedDate: &today,
	}
	db.Create(&streak)

	result, isValid, err := service.CheckStreak(1, "")
	if err != nil {
		t.Fatalf("CheckStreak failed: %v", err)
	}

	if !isValid {
		t.Error("expected streak to be valid (solved today)")
	}
	if result.CurrentStreak != 5 {
		t.Errorf("expected current streak 5, got %d", result.CurrentStreak)
	}
}

func TestCheckStreak_ValidYesterday(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	loc := time.UTC
	yesterday := service.truncateToDate(time.Now().In(loc), loc).AddDate(0, 0, -1)
	streak := models.Streak{
		UserID:         1,
		CurrentStreak:  5,
		LongestStreak:  5,
		LastSolvedDate: &yesterday,
	}
	db.Create(&streak)

	_, isValid, err := service.CheckStreak(1, "")
	if err != nil {
		t.Fatalf("CheckStreak failed: %v", err)
	}

	if !isValid {
		t.Error("expected streak to be valid (solved yesterday)")
	}
}

func TestCheckStreak_Broken(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	loc := time.UTC
	twoDaysAgo := service.truncateToDate(time.Now().In(loc), loc).AddDate(0, 0, -2)
	streak := models.Streak{
		UserID:         1,
		CurrentStreak:  5,
		LongestStreak:  5,
		LastSolvedDate: &twoDaysAgo,
	}
	db.Create(&streak)

	result, isValid, err := service.CheckStreak(1, "")
	if err != nil {
		t.Fatalf("CheckStreak failed: %v", err)
	}

	if isValid {
		t.Error("expected streak to be invalid (not solved in 2 days)")
	}
	if result.CurrentStreak != 0 {
		t.Errorf("expected current streak to be reset to 0, got %d", result.CurrentStreak)
	}
}

func TestCheckStreak_NonExistent(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	result, isValid, err := service.CheckStreak(999, "")
	if err != nil {
		t.Fatalf("CheckStreak failed: %v", err)
	}

	if result != nil {
		t.Error("expected nil result for non-existent user")
	}
	if isValid {
		t.Error("expected isValid to be false for non-existent user")
	}
}

func TestResetStreak(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Create a streak
	streak := models.Streak{
		UserID:        1,
		CurrentStreak: 10,
		LongestStreak: 10,
	}
	db.Create(&streak)

	err := service.ResetStreak(1)
	if err != nil {
		t.Fatalf("ResetStreak failed: %v", err)
	}

	// Verify reset
	result, _ := service.GetUserStreak(1)
	if result.CurrentStreak != 0 {
		t.Errorf("expected streak to be 0 after reset, got %d", result.CurrentStreak)
	}
	if result.LongestStreak != 10 {
		t.Errorf("expected longest streak to remain 10, got %d", result.LongestStreak)
	}
}

func TestGetStreakStats(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Create various streaks
	streaks := []models.Streak{
		{UserID: 1, CurrentStreak: 5, LongestStreak: 10},
		{UserID: 2, CurrentStreak: 10, LongestStreak: 15},
		{UserID: 3, CurrentStreak: 0, LongestStreak: 5}, // Broken streak
		{UserID: 4, CurrentStreak: 50, LongestStreak: 50},
	}
	for _, s := range streaks {
		db.Create(&s)
	}

	stats, err := service.GetStreakStats()
	if err != nil {
		t.Fatalf("GetStreakStats failed: %v", err)
	}

	if stats.TotalUsers != 4 {
		t.Errorf("expected 4 total users, got %d", stats.TotalUsers)
	}
	if stats.UsersWithStreak != 3 {
		t.Errorf("expected 3 users with active streak, got %d", stats.UsersWithStreak)
	}
	if stats.MaxActiveStreak != 50 {
		t.Errorf("expected max streak 50, got %d", stats.MaxActiveStreak)
	}
}

func TestGetTopStreaks(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	streaks := []models.Streak{
		{UserID: 1, CurrentStreak: 5, LongestStreak: 10},
		{UserID: 2, CurrentStreak: 15, LongestStreak: 20},
		{UserID: 3, CurrentStreak: 10, LongestStreak: 15},
		{UserID: 4, CurrentStreak: 0, LongestStreak: 50}, // Broken, shouldn't appear
	}
	for _, s := range streaks {
		db.Create(&s)
	}

	top, err := service.GetTopStreaks(3)
	if err != nil {
		t.Fatalf("GetTopStreaks failed: %v", err)
	}

	if len(top) != 3 {
		t.Errorf("expected 3 top streaks, got %d", len(top))
	}

	// Should be ordered by current streak descending
	if top[0].UserID != 2 || top[0].CurrentStreak != 15 {
		t.Errorf("expected user 2 with streak 15 at top, got user %d with streak %d",
			top[0].UserID, top[0].CurrentStreak)
	}
}

func TestInitializeStreak(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	err := service.InitializeStreak(1)
	if err != nil {
		t.Fatalf("InitializeStreak failed: %v", err)
	}

	streak, _ := service.GetUserStreak(1)
	if streak == nil {
		t.Fatal("expected streak to be created")
	}
	if streak.CurrentStreak != 0 {
		t.Errorf("expected initial streak 0, got %d", streak.CurrentStreak)
	}
}

func TestInitializeStreak_RaceCondition(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Initialize twice (should not error due to upsert)
	err := service.InitializeStreak(1)
	if err != nil {
		t.Fatalf("first InitializeStreak failed: %v", err)
	}

	err = service.InitializeStreak(1)
	if err != nil {
		t.Fatalf("second InitializeStreak failed: %v", err)
	}

	// Should still only have one record
	var count int64
	db.Model(&models.Streak{}).Where("user_id = ?", 1).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 streak record, got %d", count)
	}
}

func TestDaysUntilStreakBreaks_SolvedToday(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	loc := time.UTC
	today := service.truncateToDate(time.Now().In(loc), loc)
	streak := &models.Streak{
		UserID:         1,
		CurrentStreak:  5,
		LastSolvedDate: &today,
	}

	days := service.DaysUntilStreakBreaks(streak, "")
	if days != 2 {
		t.Errorf("expected 2 days (can skip today), got %d", days)
	}
}

func TestDaysUntilStreakBreaks_SolvedYesterday(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	loc := time.UTC
	yesterday := service.truncateToDate(time.Now().In(loc), loc).AddDate(0, 0, -1)
	streak := &models.Streak{
		UserID:         1,
		CurrentStreak:  5,
		LastSolvedDate: &yesterday,
	}

	days := service.DaysUntilStreakBreaks(streak, "")
	if days != 1 {
		t.Errorf("expected 1 day (must solve today), got %d", days)
	}
}

func TestDaysUntilStreakBreaks_AlreadyBroken(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	loc := time.UTC
	twoDaysAgo := service.truncateToDate(time.Now().In(loc), loc).AddDate(0, 0, -2)
	streak := &models.Streak{
		UserID:         1,
		CurrentStreak:  5,
		LastSolvedDate: &twoDaysAgo,
	}

	days := service.DaysUntilStreakBreaks(streak, "")
	if days != 0 {
		t.Errorf("expected 0 days (already broken), got %d", days)
	}
}

func TestDaysUntilStreakBreaks_NilStreak(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	days := service.DaysUntilStreakBreaks(nil, "")
	if days != 0 {
		t.Errorf("expected 0 days for nil streak, got %d", days)
	}
}

func TestDaysUntilStreakBreaks_ZeroStreak(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	streak := &models.Streak{
		UserID:        1,
		CurrentStreak: 0,
	}

	days := service.DaysUntilStreakBreaks(streak, "")
	if days != 0 {
		t.Errorf("expected 0 days for zero streak, got %d", days)
	}
}

// TestTimezone_DateLineScenario tests handling across the international date line
func TestTimezone_DateLineScenario(t *testing.T) {
	db := setupStreakTestDB(t)

	// Use a timezone ahead of UTC (e.g., Samoa UTC+13)
	samoaLoc, err := time.LoadLocation("Pacific/Apia")
	if err != nil {
		t.Skip("Pacific/Apia timezone not available")
	}

	service := NewStreakServiceWithTimezone(db, samoaLoc)

	update, err := service.UpdateStreak(1, "Pacific/Apia")
	if err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}

	if update.CurrentStreak != 1 {
		t.Errorf("expected streak 1, got %d", update.CurrentStreak)
	}
}

// TestTimezone_MidnightCrossing tests behavior around midnight in user's timezone
func TestTimezone_MidnightCrossing(t *testing.T) {
	db := setupStreakTestDB(t)
	service := NewStreakService(db)

	// Create streak solved at 11:59 PM yesterday in UTC
	loc := time.UTC
	yesterday := service.truncateToDate(time.Now().In(loc), loc).AddDate(0, 0, -1)
	streak := models.Streak{
		UserID:         1,
		CurrentStreak:  5,
		LongestStreak:  5,
		LastSolvedDate: &yesterday,
	}
	db.Create(&streak)

	// Check streak - should still be valid (solved yesterday)
	_, isValid, err := service.CheckStreak(1, "")
	if err != nil {
		t.Fatalf("CheckStreak failed: %v", err)
	}

	if !isValid {
		t.Error("expected streak to be valid when solved yesterday")
	}
}
