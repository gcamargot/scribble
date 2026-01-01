package services

import (
	"testing"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Mock submission model for testing
type TestSubmission struct {
	ID              uint    `gorm:"primaryKey"`
	UserID          uint    `gorm:"not null"`
	ProblemID       uint    `gorm:"not null"`
	Status          string  `gorm:"not null"`
	ExecutionTimeMs *int    `gorm:"column:execution_time_ms"`
	MemoryUsedKb    *int    `gorm:"column:memory_used_kb"`
}

func (TestSubmission) TableName() string {
	return "submissions"
}

// Mock streak model for testing
type TestStreak struct {
	ID            uint `gorm:"primaryKey"`
	UserID        uint `gorm:"not null"`
	CurrentStreak int  `gorm:"default:0"`
	LongestStreak int  `gorm:"default:0"`
}

func (TestStreak) TableName() string {
	return "streaks"
}

// Mock user model for testing
type TestUser struct {
	ID        uint    `gorm:"primaryKey"`
	Username  string  `gorm:"not null"`
	AvatarURL *string
}

func (TestUser) TableName() string {
	return "users"
}

// setupLeaderboardTestDB creates an in-memory SQLite database for testing
func setupLeaderboardTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Migrate all required tables
	err = db.AutoMigrate(
		&models.LeaderboardEntry{},
		&TestSubmission{},
		&TestStreak{},
		&TestUser{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test tables: %v", err)
	}

	return db
}

func intPtr(i int) *int {
	return &i
}

func TestComputeFastestAvg(t *testing.T) {
	db := setupLeaderboardTestDB(t)
	service := NewLeaderboardService(db)

	// Create test submissions
	submissions := []TestSubmission{
		{UserID: 1, ProblemID: 1, Status: "accepted", ExecutionTimeMs: intPtr(100)},
		{UserID: 1, ProblemID: 2, Status: "accepted", ExecutionTimeMs: intPtr(200)}, // User 1 avg: 150
		{UserID: 2, ProblemID: 1, Status: "accepted", ExecutionTimeMs: intPtr(50)},
		{UserID: 2, ProblemID: 2, Status: "accepted", ExecutionTimeMs: intPtr(50)},  // User 2 avg: 50 (faster)
		{UserID: 3, ProblemID: 1, Status: "wrong_answer", ExecutionTimeMs: intPtr(10)}, // Not counted - wrong answer
	}
	for _, s := range submissions {
		db.Create(&s)
	}

	result, err := service.ComputeLeaderboard(models.MetricFastestAvg)
	if err != nil {
		t.Fatalf("ComputeLeaderboard failed: %v", err)
	}

	if result.MetricType != models.MetricFastestAvg {
		t.Errorf("expected metric type %s, got %s", models.MetricFastestAvg, result.MetricType)
	}

	if result.EntriesUpdated != 2 {
		t.Errorf("expected 2 entries updated, got %d", result.EntriesUpdated)
	}

	// Verify rankings in cache
	var entries []models.LeaderboardEntry
	db.Where("metric_type = ?", models.MetricFastestAvg).Order("rank ASC").Find(&entries)

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// User 2 should be rank 1 (faster avg)
	if entries[0].UserID != 2 || entries[0].Rank != 1 {
		t.Errorf("expected user 2 at rank 1, got user %d at rank %d", entries[0].UserID, entries[0].Rank)
	}

	// User 1 should be rank 2
	if entries[1].UserID != 1 || entries[1].Rank != 2 {
		t.Errorf("expected user 1 at rank 2, got user %d at rank %d", entries[1].UserID, entries[1].Rank)
	}
}

func TestComputeLowestMemoryAvg(t *testing.T) {
	db := setupLeaderboardTestDB(t)
	service := NewLeaderboardService(db)

	// Create test submissions
	submissions := []TestSubmission{
		{UserID: 1, ProblemID: 1, Status: "accepted", MemoryUsedKb: intPtr(1000)},
		{UserID: 1, ProblemID: 2, Status: "accepted", MemoryUsedKb: intPtr(2000)}, // User 1 avg: 1500
		{UserID: 2, ProblemID: 1, Status: "accepted", MemoryUsedKb: intPtr(500)},
		{UserID: 2, ProblemID: 2, Status: "accepted", MemoryUsedKb: intPtr(500)},  // User 2 avg: 500 (lower)
	}
	for _, s := range submissions {
		db.Create(&s)
	}

	result, err := service.ComputeLeaderboard(models.MetricLowestMemoryAvg)
	if err != nil {
		t.Fatalf("ComputeLeaderboard failed: %v", err)
	}

	if result.EntriesUpdated != 2 {
		t.Errorf("expected 2 entries updated, got %d", result.EntriesUpdated)
	}

	// Verify rankings
	var entries []models.LeaderboardEntry
	db.Where("metric_type = ?", models.MetricLowestMemoryAvg).Order("rank ASC").Find(&entries)

	// User 2 should be rank 1 (lower memory)
	if entries[0].UserID != 2 || entries[0].Rank != 1 {
		t.Errorf("expected user 2 at rank 1, got user %d at rank %d", entries[0].UserID, entries[0].Rank)
	}
}

func TestComputeProblemsSolved(t *testing.T) {
	db := setupLeaderboardTestDB(t)
	service := NewLeaderboardService(db)

	// Create test submissions
	submissions := []TestSubmission{
		// User 1: solved problems 1, 2, 3 (3 unique)
		{UserID: 1, ProblemID: 1, Status: "accepted"},
		{UserID: 1, ProblemID: 2, Status: "accepted"},
		{UserID: 1, ProblemID: 3, Status: "accepted"},
		// User 2: solved problems 1, 2 (2 unique)
		{UserID: 2, ProblemID: 1, Status: "accepted"},
		{UserID: 2, ProblemID: 2, Status: "accepted"},
		{UserID: 2, ProblemID: 2, Status: "accepted"}, // Duplicate problem - shouldn't count twice
		// User 3: no accepted solutions
		{UserID: 3, ProblemID: 1, Status: "wrong_answer"},
	}
	for _, s := range submissions {
		db.Create(&s)
	}

	result, err := service.ComputeLeaderboard(models.MetricProblemsSolved)
	if err != nil {
		t.Fatalf("ComputeLeaderboard failed: %v", err)
	}

	if result.EntriesUpdated != 2 {
		t.Errorf("expected 2 entries updated, got %d", result.EntriesUpdated)
	}

	// Verify rankings
	var entries []models.LeaderboardEntry
	db.Where("metric_type = ?", models.MetricProblemsSolved).Order("rank ASC").Find(&entries)

	// User 1 should be rank 1 (more problems solved)
	if entries[0].UserID != 1 || entries[0].Rank != 1 {
		t.Errorf("expected user 1 at rank 1, got user %d at rank %d", entries[0].UserID, entries[0].Rank)
	}

	if entries[0].MetricValue != 3 {
		t.Errorf("expected metric value 3 for user 1, got %f", entries[0].MetricValue)
	}

	// User 2 should be rank 2
	if entries[1].UserID != 2 || entries[1].Rank != 2 {
		t.Errorf("expected user 2 at rank 2, got user %d at rank %d", entries[1].UserID, entries[1].Rank)
	}

	if entries[1].MetricValue != 2 {
		t.Errorf("expected metric value 2 for user 2, got %f", entries[1].MetricValue)
	}
}

func TestComputeLongestStreak(t *testing.T) {
	db := setupLeaderboardTestDB(t)
	service := NewLeaderboardService(db)

	// Create test streaks
	streaks := []TestStreak{
		{UserID: 1, CurrentStreak: 5, LongestStreak: 10},
		{UserID: 2, CurrentStreak: 3, LongestStreak: 15}, // User 2 has longest streak
		{UserID: 3, CurrentStreak: 0, LongestStreak: 0},  // User 3 has no streak
	}
	for _, s := range streaks {
		db.Create(&s)
	}

	result, err := service.ComputeLeaderboard(models.MetricLongestStreak)
	if err != nil {
		t.Fatalf("ComputeLeaderboard failed: %v", err)
	}

	// User 3 has 0 streak so not counted
	if result.EntriesUpdated != 2 {
		t.Errorf("expected 2 entries updated, got %d", result.EntriesUpdated)
	}

	// Verify rankings
	var entries []models.LeaderboardEntry
	db.Where("metric_type = ?", models.MetricLongestStreak).Order("rank ASC").Find(&entries)

	// User 2 should be rank 1 (longest streak)
	if entries[0].UserID != 2 || entries[0].Rank != 1 {
		t.Errorf("expected user 2 at rank 1, got user %d at rank %d", entries[0].UserID, entries[0].Rank)
	}

	if entries[0].MetricValue != 15 {
		t.Errorf("expected metric value 15, got %f", entries[0].MetricValue)
	}
}

func TestComputeAllLeaderboards(t *testing.T) {
	db := setupLeaderboardTestDB(t)
	service := NewLeaderboardService(db)

	// Create minimal test data
	db.Create(&TestSubmission{UserID: 1, ProblemID: 1, Status: "accepted", ExecutionTimeMs: intPtr(100), MemoryUsedKb: intPtr(1000)})
	db.Create(&TestStreak{UserID: 1, CurrentStreak: 1, LongestStreak: 5})

	results, err := service.ComputeAllLeaderboards()
	if err != nil {
		t.Fatalf("ComputeAllLeaderboards failed: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}

	// Verify all metric types are present
	metricTypes := make(map[models.MetricType]bool)
	for _, r := range results {
		metricTypes[r.MetricType] = true
	}

	for _, expected := range models.AllMetricTypes() {
		if !metricTypes[expected] {
			t.Errorf("missing metric type: %s", expected)
		}
	}
}

func TestGetLeaderboard(t *testing.T) {
	db := setupLeaderboardTestDB(t)
	service := NewLeaderboardService(db)

	// Create test users
	users := []TestUser{
		{ID: 1, Username: "alice"},
		{ID: 2, Username: "bob"},
		{ID: 3, Username: "charlie"},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// Create leaderboard entries
	now := time.Now()
	entries := []models.LeaderboardEntry{
		{UserID: 1, MetricType: models.MetricProblemsSolved, MetricValue: 10, Rank: 1, ComputedAt: now},
		{UserID: 2, MetricType: models.MetricProblemsSolved, MetricValue: 8, Rank: 2, ComputedAt: now},
		{UserID: 3, MetricType: models.MetricProblemsSolved, MetricValue: 5, Rank: 3, ComputedAt: now},
	}
	for _, e := range entries {
		db.Create(&e)
	}

	// Test pagination
	page, err := service.GetLeaderboard(models.MetricProblemsSolved, 1, 2)
	if err != nil {
		t.Fatalf("GetLeaderboard failed: %v", err)
	}

	if page.Total != 3 {
		t.Errorf("expected total 3, got %d", page.Total)
	}

	if page.TotalPages != 2 {
		t.Errorf("expected 2 total pages, got %d", page.TotalPages)
	}

	if len(page.Entries) != 2 {
		t.Errorf("expected 2 entries on page 1, got %d", len(page.Entries))
	}

	if page.Entries[0].Username != "alice" {
		t.Errorf("expected first entry to be alice, got %s", page.Entries[0].Username)
	}
}

func TestGetUserRank(t *testing.T) {
	db := setupLeaderboardTestDB(t)
	service := NewLeaderboardService(db)

	// Create leaderboard entry
	entry := models.LeaderboardEntry{
		UserID:      1,
		MetricType:  models.MetricProblemsSolved,
		MetricValue: 10,
		Rank:        5,
		ComputedAt:  time.Now(),
	}
	db.Create(&entry)

	// Get existing rank
	result, err := service.GetUserRank(1, models.MetricProblemsSolved)
	if err != nil {
		t.Fatalf("GetUserRank failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Rank != 5 {
		t.Errorf("expected rank 5, got %d", result.Rank)
	}

	// Get non-existing rank
	result, err = service.GetUserRank(999, models.MetricProblemsSolved)
	if err != nil {
		t.Fatalf("GetUserRank failed: %v", err)
	}

	if result != nil {
		t.Error("expected nil for non-existing user")
	}
}

func TestGetUserAllRanks(t *testing.T) {
	db := setupLeaderboardTestDB(t)
	service := NewLeaderboardService(db)

	// Create multiple leaderboard entries for user 1
	now := time.Now()
	entries := []models.LeaderboardEntry{
		{UserID: 1, MetricType: models.MetricProblemsSolved, MetricValue: 10, Rank: 1, ComputedAt: now},
		{UserID: 1, MetricType: models.MetricFastestAvg, MetricValue: 50.5, Rank: 3, ComputedAt: now},
		{UserID: 2, MetricType: models.MetricProblemsSolved, MetricValue: 5, Rank: 2, ComputedAt: now}, // Different user
	}
	for _, e := range entries {
		db.Create(&e)
	}

	result, err := service.GetUserAllRanks(1)
	if err != nil {
		t.Fatalf("GetUserAllRanks failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 entries for user 1, got %d", len(result))
	}
}

func TestComputeLeaderboard_UnknownMetricType(t *testing.T) {
	db := setupLeaderboardTestDB(t)
	service := NewLeaderboardService(db)

	_, err := service.ComputeLeaderboard(models.MetricType("unknown"))
	if err == nil {
		t.Error("expected error for unknown metric type")
	}
}

func TestAllMetricTypes(t *testing.T) {
	types := models.AllMetricTypes()

	if len(types) != 4 {
		t.Errorf("expected 4 metric types, got %d", len(types))
	}

	expected := map[models.MetricType]bool{
		models.MetricFastestAvg:      false,
		models.MetricLowestMemoryAvg: false,
		models.MetricProblemsSolved:  false,
		models.MetricLongestStreak:   false,
	}

	for _, mt := range types {
		if _, ok := expected[mt]; !ok {
			t.Errorf("unexpected metric type: %s", mt)
		}
		expected[mt] = true
	}

	for mt, found := range expected {
		if !found {
			t.Errorf("missing metric type: %s", mt)
		}
	}
}
