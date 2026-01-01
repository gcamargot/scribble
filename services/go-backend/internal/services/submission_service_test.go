package services

import (
	"testing"
	"time"

	"github.com/nahtao97/scribble/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupSubmissionTestDB creates an in-memory SQLite database for testing
func setupSubmissionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Migrate all required tables
	err = db.AutoMigrate(
		&models.Problem{},
		&models.Submission{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test tables: %v", err)
	}

	return db
}

// createTestProblem creates a test problem in the database
func createTestProblem(db *gorm.DB) *models.Problem {
	problem := &models.Problem{
		Title:       "Test Problem",
		Slug:        "test-problem",
		Difficulty:  "easy",
		Description: "Test description",
	}
	db.Create(problem)
	return problem
}

func TestNewSubmissionService(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)

	if service == nil {
		t.Fatal("expected service to be created")
	}
}

func TestCreateSubmission(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem := createTestProblem(db)

	submission := &models.Submission{
		UserID:    1,
		ProblemID: problem.ID,
		Language:  "python",
		Code:      "print('hello')",
		Status:    models.StatusAccepted,
	}

	err := service.CreateSubmission(submission)
	if err != nil {
		t.Fatalf("CreateSubmission failed: %v", err)
	}

	if submission.ID == 0 {
		t.Error("expected submission to have an ID after creation")
	}

	// Verify in database
	var dbSubmission models.Submission
	db.First(&dbSubmission, submission.ID)
	if dbSubmission.Code != "print('hello')" {
		t.Errorf("expected code to be saved, got '%s'", dbSubmission.Code)
	}
}

func TestGetSubmissionByID_Found(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem := createTestProblem(db)

	submission := models.Submission{
		UserID:    1,
		ProblemID: problem.ID,
		Language:  "go",
		Code:      "fmt.Println()",
		Status:    models.StatusAccepted,
	}
	db.Create(&submission)

	result, err := service.GetSubmissionByID(submission.ID)
	if err != nil {
		t.Fatalf("GetSubmissionByID failed: %v", err)
	}

	if result.Language != "go" {
		t.Errorf("expected language 'go', got '%s'", result.Language)
	}

	// Verify problem is preloaded
	if result.Problem.Title != "Test Problem" {
		t.Errorf("expected preloaded problem, got '%s'", result.Problem.Title)
	}
}

func TestGetSubmissionByID_NotFound(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)

	_, err := service.GetSubmissionByID(999)
	if err == nil {
		t.Error("expected error for non-existent submission")
	}
}

func TestGetSubmissionWithCode(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem := createTestProblem(db)

	submission := models.Submission{
		UserID:    1,
		ProblemID: problem.ID,
		Language:  "python",
		Code:      "def solution(): pass",
		Status:    models.StatusAccepted,
	}
	db.Create(&submission)

	result, err := service.GetSubmissionWithCode(submission.ID)
	if err != nil {
		t.Fatalf("GetSubmissionWithCode failed: %v", err)
	}

	// Verify code is included
	if result.Code != "def solution(): pass" {
		t.Errorf("expected code to be included, got '%s'", result.Code)
	}
}

func TestGetSubmissionWithCode_NotFound(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)

	_, err := service.GetSubmissionWithCode(999)
	if err == nil {
		t.Error("expected error for non-existent submission")
	}
}

func TestGetUserSubmissionHistory_Basic(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem := createTestProblem(db)

	// Create 5 submissions for user 1
	for i := 0; i < 5; i++ {
		submission := models.Submission{
			UserID:      1,
			ProblemID:   problem.ID,
			Language:    "python",
			Code:        "code",
			Status:      models.StatusAccepted,
			SubmittedAt: time.Now().Add(time.Duration(-i) * time.Hour),
		}
		db.Create(&submission)
	}

	params := SubmissionHistoryParams{
		UserID:   1,
		Page:     1,
		PageSize: 10,
	}

	result, err := service.GetUserSubmissionHistory(params)
	if err != nil {
		t.Fatalf("GetUserSubmissionHistory failed: %v", err)
	}

	if result.Total != 5 {
		t.Errorf("expected 5 total submissions, got %d", result.Total)
	}
	if len(result.Submissions) != 5 {
		t.Errorf("expected 5 submissions, got %d", len(result.Submissions))
	}
}

func TestGetUserSubmissionHistory_Pagination(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem := createTestProblem(db)

	// Create 25 submissions
	for i := 0; i < 25; i++ {
		submission := models.Submission{
			UserID:      1,
			ProblemID:   problem.ID,
			Language:    "python",
			Code:        "code",
			Status:      models.StatusAccepted,
			SubmittedAt: time.Now().Add(time.Duration(-i) * time.Minute),
		}
		db.Create(&submission)
	}

	// Test first page
	params := SubmissionHistoryParams{
		UserID:   1,
		Page:     1,
		PageSize: 10,
	}

	result, err := service.GetUserSubmissionHistory(params)
	if err != nil {
		t.Fatalf("GetUserSubmissionHistory failed: %v", err)
	}

	if result.Total != 25 {
		t.Errorf("expected 25 total, got %d", result.Total)
	}
	if result.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", result.TotalPages)
	}
	if len(result.Submissions) != 10 {
		t.Errorf("expected 10 submissions on page 1, got %d", len(result.Submissions))
	}

	// Test second page
	params.Page = 2
	result, err = service.GetUserSubmissionHistory(params)
	if err != nil {
		t.Fatalf("GetUserSubmissionHistory page 2 failed: %v", err)
	}
	if len(result.Submissions) != 10 {
		t.Errorf("expected 10 submissions on page 2, got %d", len(result.Submissions))
	}

	// Test last page
	params.Page = 3
	result, err = service.GetUserSubmissionHistory(params)
	if err != nil {
		t.Fatalf("GetUserSubmissionHistory page 3 failed: %v", err)
	}
	if len(result.Submissions) != 5 {
		t.Errorf("expected 5 submissions on page 3, got %d", len(result.Submissions))
	}
}

func TestGetUserSubmissionHistory_FilterByStatus(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem := createTestProblem(db)

	// Create submissions with different statuses
	statuses := []models.SubmissionStatus{
		models.StatusAccepted,
		models.StatusAccepted,
		models.StatusWrongAnswer,
		models.StatusRuntimeError,
	}

	for _, status := range statuses {
		submission := models.Submission{
			UserID:    1,
			ProblemID: problem.ID,
			Language:  "python",
			Code:      "code",
			Status:    status,
		}
		db.Create(&submission)
	}

	params := SubmissionHistoryParams{
		UserID: 1,
		Status: "accepted",
	}

	result, err := service.GetUserSubmissionHistory(params)
	if err != nil {
		t.Fatalf("GetUserSubmissionHistory failed: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected 2 accepted submissions, got %d", result.Total)
	}
}

func TestGetUserSubmissionHistory_FilterByLanguage(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem := createTestProblem(db)

	// Create submissions with different languages
	languages := []string{"python", "python", "go", "java"}

	for _, lang := range languages {
		submission := models.Submission{
			UserID:    1,
			ProblemID: problem.ID,
			Language:  lang,
			Code:      "code",
			Status:    models.StatusAccepted,
		}
		db.Create(&submission)
	}

	params := SubmissionHistoryParams{
		UserID:   1,
		Language: "python",
	}

	result, err := service.GetUserSubmissionHistory(params)
	if err != nil {
		t.Fatalf("GetUserSubmissionHistory failed: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected 2 python submissions, got %d", result.Total)
	}
}

func TestGetUserSubmissionHistory_FilterByProblem(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)

	// Create two problems
	problem1 := createTestProblem(db)
	problem2 := &models.Problem{
		Title:       "Problem 2",
		Slug:        "problem-2",
		Difficulty:  "medium",
		Description: "Another problem",
	}
	db.Create(problem2)

	// Create submissions for both problems
	db.Create(&models.Submission{UserID: 1, ProblemID: problem1.ID, Language: "python", Code: "1", Status: models.StatusAccepted})
	db.Create(&models.Submission{UserID: 1, ProblemID: problem1.ID, Language: "python", Code: "2", Status: models.StatusAccepted})
	db.Create(&models.Submission{UserID: 1, ProblemID: problem2.ID, Language: "python", Code: "3", Status: models.StatusAccepted})

	params := SubmissionHistoryParams{
		UserID:    1,
		ProblemID: &problem1.ID,
	}

	result, err := service.GetUserSubmissionHistory(params)
	if err != nil {
		t.Fatalf("GetUserSubmissionHistory failed: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected 2 submissions for problem 1, got %d", result.Total)
	}
}

func TestGetUserSubmissionHistory_Defaults(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)

	// Test with invalid page and page size
	params := SubmissionHistoryParams{
		UserID:   1,
		Page:     0,  // Invalid - should default to 1
		PageSize: -1, // Invalid - should default to 20
	}

	result, err := service.GetUserSubmissionHistory(params)
	if err != nil {
		t.Fatalf("GetUserSubmissionHistory failed: %v", err)
	}

	if result.Page != 1 {
		t.Errorf("expected page to default to 1, got %d", result.Page)
	}
	if result.PageSize != 20 {
		t.Errorf("expected page size to default to 20, got %d", result.PageSize)
	}
}

func TestGetUserSubmissionHistory_MaxPageSize(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)

	params := SubmissionHistoryParams{
		UserID:   1,
		PageSize: 500, // Exceeds max - should default to 20
	}

	result, err := service.GetUserSubmissionHistory(params)
	if err != nil {
		t.Fatalf("GetUserSubmissionHistory failed: %v", err)
	}

	if result.PageSize != 20 {
		t.Errorf("expected page size to default to 20 when exceeding max, got %d", result.PageSize)
	}
}

func TestGetSubmissionsByUserAndProblem(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem := createTestProblem(db)

	// Create submissions for user 1 on problem
	for i := 0; i < 3; i++ {
		db.Create(&models.Submission{
			UserID:      1,
			ProblemID:   problem.ID,
			Language:    "python",
			Code:        "code",
			Status:      models.StatusAccepted,
			SubmittedAt: time.Now().Add(time.Duration(-i) * time.Hour),
		})
	}

	// Create submission for different user
	db.Create(&models.Submission{
		UserID:    2,
		ProblemID: problem.ID,
		Language:  "python",
		Code:      "code",
		Status:    models.StatusAccepted,
	})

	result, err := service.GetSubmissionsByUserAndProblem(1, problem.ID)
	if err != nil {
		t.Fatalf("GetSubmissionsByUserAndProblem failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 submissions for user 1, got %d", len(result))
	}
}

func TestGetUserSubmissionStats(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem1 := createTestProblem(db)
	problem2 := &models.Problem{
		Title:       "Problem 2",
		Slug:        "problem-2",
		Difficulty:  "medium",
		Description: "Another problem",
	}
	db.Create(problem2)

	execTime1 := 100
	execTime2 := 200
	memory1 := 1000
	memory2 := 2000

	// Create various submissions
	submissions := []models.Submission{
		{UserID: 1, ProblemID: problem1.ID, Language: "python", Code: "1", Status: models.StatusAccepted, ExecutionTimeMs: &execTime1, MemoryUsedKb: &memory1},
		{UserID: 1, ProblemID: problem1.ID, Language: "python", Code: "2", Status: models.StatusWrongAnswer},
		{UserID: 1, ProblemID: problem2.ID, Language: "python", Code: "3", Status: models.StatusAccepted, ExecutionTimeMs: &execTime2, MemoryUsedKb: &memory2},
		{UserID: 1, ProblemID: problem2.ID, Language: "python", Code: "4", Status: models.StatusRuntimeError},
	}

	for _, s := range submissions {
		db.Create(&s)
	}

	stats, err := service.GetUserSubmissionStats(1)
	if err != nil {
		t.Fatalf("GetUserSubmissionStats failed: %v", err)
	}

	if stats.TotalSubmissions != 4 {
		t.Errorf("expected 4 total submissions, got %d", stats.TotalSubmissions)
	}
	if stats.AcceptedSubmissions != 2 {
		t.Errorf("expected 2 accepted submissions, got %d", stats.AcceptedSubmissions)
	}
	if stats.ProblemsSolved != 2 {
		t.Errorf("expected 2 problems solved, got %d", stats.ProblemsSolved)
	}
	if stats.AcceptanceRate != 50.0 {
		t.Errorf("expected 50%% acceptance rate, got %.2f%%", stats.AcceptanceRate)
	}

	// Check average execution time (100 + 200) / 2 = 150
	if stats.AvgExecutionTimeMs == nil || *stats.AvgExecutionTimeMs != 150 {
		t.Errorf("expected avg execution time 150, got %v", stats.AvgExecutionTimeMs)
	}

	// Check average memory (1000 + 2000) / 2 = 1500
	if stats.AvgMemoryUsedKb == nil || *stats.AvgMemoryUsedKb != 1500 {
		t.Errorf("expected avg memory 1500, got %v", stats.AvgMemoryUsedKb)
	}
}

func TestGetUserSubmissionStats_NoSubmissions(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)

	stats, err := service.GetUserSubmissionStats(999)
	if err != nil {
		t.Fatalf("GetUserSubmissionStats failed: %v", err)
	}

	if stats.TotalSubmissions != 0 {
		t.Errorf("expected 0 total submissions, got %d", stats.TotalSubmissions)
	}
	if stats.AcceptanceRate != 0 {
		t.Errorf("expected 0%% acceptance rate, got %.2f%%", stats.AcceptanceRate)
	}
	if stats.AvgExecutionTimeMs != nil {
		t.Error("expected nil avg execution time for no submissions")
	}
}

func TestGetUserSubmissionStats_NoDuplicateProblemCount(t *testing.T) {
	db := setupSubmissionTestDB(t)
	service := NewSubmissionService(db)
	problem := createTestProblem(db)

	// Create multiple accepted submissions for the same problem
	for i := 0; i < 5; i++ {
		db.Create(&models.Submission{
			UserID:    1,
			ProblemID: problem.ID,
			Language:  "python",
			Code:      "code",
			Status:    models.StatusAccepted,
		})
	}

	stats, err := service.GetUserSubmissionStats(1)
	if err != nil {
		t.Fatalf("GetUserSubmissionStats failed: %v", err)
	}

	// Should only count as 1 problem solved (unique problems)
	if stats.ProblemsSolved != 1 {
		t.Errorf("expected 1 problem solved (unique), got %d", stats.ProblemsSolved)
	}
}
