package services

import (
	"testing"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupProblemTestDB creates an in-memory SQLite database for testing
func setupProblemTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Migrate all required tables
	err = db.AutoMigrate(
		&models.Problem{},
		&models.TestCase{},
		&models.DailyChallenge{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test tables: %v", err)
	}

	return db
}

func TestNewProblemService(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	if service == nil {
		t.Fatal("expected service to be created")
	}
}

func TestGetProblemByID_Found(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	// Create a test problem
	problem := models.Problem{
		Title:       "Two Sum",
		Slug:        "two-sum",
		Difficulty:  "easy",
		Description: "Given an array of integers...",
	}
	db.Create(&problem)

	// Retrieve the problem
	result, err := service.GetProblemByID(problem.ID)
	if err != nil {
		t.Fatalf("GetProblemByID failed: %v", err)
	}

	if result.Title != "Two Sum" {
		t.Errorf("expected title 'Two Sum', got '%s'", result.Title)
	}
	if result.Difficulty != "easy" {
		t.Errorf("expected difficulty 'easy', got '%s'", result.Difficulty)
	}
}

func TestGetProblemByID_NotFound(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	_, err := service.GetProblemByID(999)
	if err == nil {
		t.Error("expected error for non-existent problem")
	}
}

func TestGetTestCasesByProblemID_All(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	// Create a problem
	problem := models.Problem{
		Title:       "Test Problem",
		Slug:        "test-problem",
		Difficulty:  "easy",
		Description: "Test description",
	}
	db.Create(&problem)

	// Create test cases - mix of sample and hidden
	testCases := []models.TestCase{
		{ProblemID: problem.ID, Input: "[1,2]", ExpectedOutput: "3", IsSample: true},
		{ProblemID: problem.ID, Input: "[3,4]", ExpectedOutput: "7", IsSample: true},
		{ProblemID: problem.ID, Input: "[5,6]", ExpectedOutput: "11", IsSample: false}, // Hidden
	}
	for _, tc := range testCases {
		db.Create(&tc)
	}

	// Get all test cases
	result, err := service.GetTestCasesByProblemID(problem.ID, false)
	if err != nil {
		t.Fatalf("GetTestCasesByProblemID failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 test cases, got %d", len(result))
	}
}

func TestGetTestCasesByProblemID_SampleOnly(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	// Create a problem
	problem := models.Problem{
		Title:       "Test Problem",
		Slug:        "test-problem",
		Difficulty:  "easy",
		Description: "Test description",
	}
	db.Create(&problem)

	// Create test cases - mix of sample and hidden
	testCases := []models.TestCase{
		{ProblemID: problem.ID, Input: "[1,2]", ExpectedOutput: "3", IsSample: true},
		{ProblemID: problem.ID, Input: "[3,4]", ExpectedOutput: "7", IsSample: true},
		{ProblemID: problem.ID, Input: "[5,6]", ExpectedOutput: "11", IsSample: false}, // Hidden
	}
	for _, tc := range testCases {
		db.Create(&tc)
	}

	// Get only sample test cases
	result, err := service.GetTestCasesByProblemID(problem.ID, true)
	if err != nil {
		t.Fatalf("GetTestCasesByProblemID failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 sample test cases, got %d", len(result))
	}

	// Verify all are samples
	for _, tc := range result {
		if !tc.IsSample {
			t.Error("expected only sample test cases")
		}
	}
}

func TestGetTestCasesByProblemID_NoProblem(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	// Get test cases for non-existent problem
	result, err := service.GetTestCasesByProblemID(999, false)
	if err != nil {
		t.Fatalf("GetTestCasesByProblemID should not error for empty result: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 test cases, got %d", len(result))
	}
}

func TestGetTestCasesByProblemID_OrderedByID(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	problem := models.Problem{
		Title:       "Test Problem",
		Slug:        "test-problem",
		Difficulty:  "easy",
		Description: "Test description",
	}
	db.Create(&problem)

	// Create test cases (IDs will be assigned in order)
	for i := 0; i < 5; i++ {
		tc := models.TestCase{ProblemID: problem.ID, Input: "input", ExpectedOutput: "output", IsSample: true}
		db.Create(&tc)
	}

	result, err := service.GetTestCasesByProblemID(problem.ID, false)
	if err != nil {
		t.Fatalf("GetTestCasesByProblemID failed: %v", err)
	}

	// Verify ordering
	for i := 1; i < len(result); i++ {
		if result[i].ID < result[i-1].ID {
			t.Error("test cases should be ordered by ID ascending")
		}
	}
}

func TestGetDailyChallengeByDate_Found(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	// Create a problem
	problem := models.Problem{
		Title:       "Daily Problem",
		Slug:        "daily-problem",
		Difficulty:  "medium",
		Description: "Daily challenge description",
	}
	db.Create(&problem)

	// Create a daily challenge for today
	today := time.Now().UTC()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	challenge := models.DailyChallenge{
		ProblemID:     problem.ID,
		ChallengeDate: todayDate,
	}
	db.Create(&challenge)

	// Retrieve the challenge
	result, err := service.GetDailyChallengeByDate(today)
	if err != nil {
		t.Fatalf("GetDailyChallengeByDate failed: %v", err)
	}

	if result.ProblemID != problem.ID {
		t.Errorf("expected problem ID %d, got %d", problem.ID, result.ProblemID)
	}

	// Verify problem is preloaded
	if result.Problem.Title != "Daily Problem" {
		t.Errorf("expected preloaded problem title 'Daily Problem', got '%s'", result.Problem.Title)
	}
}

func TestGetDailyChallengeByDate_NotFound(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	// Try to get challenge for a date with no challenge
	someDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := service.GetDailyChallengeByDate(someDate)

	if err == nil {
		t.Error("expected error for date with no challenge")
	}
}

func TestGetDailyChallengeByDate_TruncatesTime(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	// Create a problem
	problem := models.Problem{
		Title:       "Daily Problem",
		Slug:        "daily-problem",
		Difficulty:  "medium",
		Description: "Daily challenge description",
	}
	db.Create(&problem)

	// Create a daily challenge for a specific date
	targetDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	challenge := models.DailyChallenge{
		ProblemID:     problem.ID,
		ChallengeDate: targetDate,
	}
	db.Create(&challenge)

	// Query with a time in the middle of the day - should still match
	queryTime := time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.UTC)
	result, err := service.GetDailyChallengeByDate(queryTime)
	if err != nil {
		t.Fatalf("GetDailyChallengeByDate should truncate time to date: %v", err)
	}

	if result.ProblemID != problem.ID {
		t.Errorf("expected problem ID %d, got %d", problem.ID, result.ProblemID)
	}
}

func TestGetTodaysDailyChallenge(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	// Create a problem
	problem := models.Problem{
		Title:       "Today's Problem",
		Slug:        "todays-problem",
		Difficulty:  "hard",
		Description: "Today's challenge description",
	}
	db.Create(&problem)

	// Create a daily challenge for today
	today := time.Now().UTC()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	challenge := models.DailyChallenge{
		ProblemID:     problem.ID,
		ChallengeDate: todayDate,
	}
	db.Create(&challenge)

	// Get today's challenge
	result, err := service.GetTodaysDailyChallenge()
	if err != nil {
		t.Fatalf("GetTodaysDailyChallenge failed: %v", err)
	}

	if result.ProblemID != problem.ID {
		t.Errorf("expected problem ID %d, got %d", problem.ID, result.ProblemID)
	}
}

func TestGetTodaysDailyChallenge_NotFound(t *testing.T) {
	db := setupProblemTestDB(t)
	service := NewProblemService(db)

	// No challenge created for today
	_, err := service.GetTodaysDailyChallenge()
	if err == nil {
		t.Error("expected error when no challenge exists for today")
	}
}
