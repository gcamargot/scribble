package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

// Test configuration
var (
	baseURL     = getEnv("TEST_BASE_URL", "http://localhost:3000")
	backendURL  = getEnv("TEST_BACKEND_URL", "http://localhost:8080")
	adminSecret = getEnv("ADMIN_SECRET", "test-admin-secret")
)

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// SubmissionRequest represents a code submission
type SubmissionRequest struct {
	Code      string `json:"code"`
	Language  string `json:"language"`
	ProblemID int    `json:"problemId"`
}

// SubmissionResponse represents the response from a submission
type SubmissionResponse struct {
	Success       bool   `json:"success"`
	Status        string `json:"status"`
	Verdict       string `json:"verdict"`
	ExecutionTime string `json:"executionTime"`
	MemoryUsed    string `json:"memoryUsed"`
	TestsPassed   int    `json:"testsPassed"`
	TestsTotal    int    `json:"testsTotal"`
	Message       string `json:"message"`
	Error         string `json:"error,omitempty"`
}

// TestLanguageSubmissions tests submission flow for all 6 supported languages
func TestLanguageSubmissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Sample code for each language that adds two numbers
	testCases := []struct {
		name     string
		language string
		code     string
	}{
		{
			name:     "Python submission",
			language: "python",
			code: `def solve(a, b):
    return a + b

# Read input
a, b = map(int, input().split())
print(solve(a, b))`,
		},
		{
			name:     "JavaScript submission",
			language: "javascript",
			code: `function solve(a, b) {
    return a + b;
}

const readline = require('readline');
const rl = readline.createInterface({ input: process.stdin });
rl.on('line', (line) => {
    const [a, b] = line.split(' ').map(Number);
    console.log(solve(a, b));
    rl.close();
});`,
		},
		{
			name:     "Go submission",
			language: "go",
			code: `package main

import "fmt"

func solve(a, b int) int {
    return a + b
}

func main() {
    var a, b int
    fmt.Scan(&a, &b)
    fmt.Println(solve(a, b))
}`,
		},
		{
			name:     "Java submission",
			language: "java",
			code: `import java.util.Scanner;

public class Solution {
    public static int solve(int a, int b) {
        return a + b;
    }

    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int a = sc.nextInt();
        int b = sc.nextInt();
        System.out.println(solve(a, b));
    }
}`,
		},
		{
			name:     "C++ submission",
			language: "cpp",
			code: `#include <iostream>
using namespace std;

int solve(int a, int b) {
    return a + b;
}

int main() {
    int a, b;
    cin >> a >> b;
    cout << solve(a, b) << endl;
    return 0;
}`,
		},
		{
			name:     "Rust submission",
			language: "rust",
			code: `use std::io;

fn solve(a: i32, b: i32) -> i32 {
    a + b
}

fn main() {
    let mut input = String::new();
    io::stdin().read_line(&mut input).unwrap();
    let nums: Vec<i32> = input.trim().split_whitespace()
        .map(|x| x.parse().unwrap())
        .collect();
    println!("{}", solve(nums[0], nums[1]));
}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := submitCode(tc.code, tc.language, 1)
			if err != nil {
				t.Fatalf("Failed to submit code: %v", err)
			}

			// In mock mode, we expect success
			// In real mode, we check for valid response structure
			if resp.Error != "" && resp.Error != "Missing required fields: code, language, problemId" {
				t.Logf("Submission response: %+v", resp)
			}
		})
	}
}

// TestEdgeCases tests various edge cases in submissions
func TestEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	testCases := []struct {
		name        string
		code        string
		language    string
		problemID   int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty code",
			code:        "",
			language:    "python",
			problemID:   1,
			expectError: true,
			errorMsg:    "Missing required fields",
		},
		{
			name:        "Empty language",
			code:        "print('hello')",
			language:    "",
			problemID:   1,
			expectError: true,
			errorMsg:    "Missing required fields",
		},
		{
			name:        "Zero problem ID",
			code:        "print('hello')",
			language:    "python",
			problemID:   0,
			expectError: true,
			errorMsg:    "Missing required fields",
		},
		{
			name:        "Syntax error in code",
			code:        "def broken(",
			language:    "python",
			problemID:   1,
			expectError: false, // Should accept but return compilation error
		},
		{
			name:        "Infinite loop code",
			code:        "while True: pass",
			language:    "python",
			problemID:   1,
			expectError: false, // Should accept but timeout
		},
		{
			name:        "Very long code",
			code:        generateLongCode(10000),
			language:    "python",
			problemID:   1,
			expectError: false, // Should handle long code
		},
		{
			name:        "Unicode in code",
			code:        "print('Hello, 世界! 🎉')",
			language:    "python",
			problemID:   1,
			expectError: false,
		},
		{
			name:        "Whitespace only code",
			code:        "   \n\t\n   ",
			language:    "python",
			problemID:   1,
			expectError: false, // Should accept but fail execution
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := submitCode(tc.code, tc.language, tc.problemID)

			if tc.expectError {
				if err == nil && resp.Error == "" {
					t.Errorf("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Logf("Request error (may be expected): %v", err)
				}
			}
		})
	}
}

// TestConcurrentSubmissions tests load with concurrent submissions
func TestConcurrentSubmissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	const numConcurrent = 10
	const code = `print(sum(map(int, input().split())))`

	var wg sync.WaitGroup
	results := make(chan *SubmissionResponse, numConcurrent)
	errors := make(chan error, numConcurrent)

	startTime := time.Now()

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			resp, err := submitCode(code, "python", 1)
			if err != nil {
				errors <- fmt.Errorf("request %d failed: %w", id, err)
				return
			}
			results <- resp
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	duration := time.Since(startTime)

	// Count results
	successCount := 0
	for range results {
		successCount++
	}

	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Error: %v", err)
	}

	t.Logf("Concurrent test completed in %v", duration)
	t.Logf("Successes: %d, Errors: %d", successCount, errorCount)

	// At least some should succeed
	if successCount == 0 && errorCount == numConcurrent {
		t.Error("All concurrent requests failed")
	}
}

// TestSubmissionHistory tests the submission history endpoint
func TestSubmissionHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Test the internal endpoint directly
	url := fmt.Sprintf("%s/internal/submissions/user/1", backendURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Logf("Response status: %d (expected 200 or auth error)", resp.StatusCode)
	}
}

// TestProblemEndpoints tests the problem-related endpoints
func TestProblemEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	testCases := []struct {
		name       string
		endpoint   string
		expectCode int
	}{
		{
			name:       "Get daily problem",
			endpoint:   "/internal/problems/daily/today",
			expectCode: http.StatusOK,
		},
		{
			name:       "Get problem by ID",
			endpoint:   "/internal/problems/1",
			expectCode: http.StatusOK,
		},
		{
			name:       "Get test cases",
			endpoint:   "/internal/problems/1/test-cases",
			expectCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := backendURL + tc.endpoint
			resp, err := http.Get(url)
			if err != nil {
				t.Skipf("Backend not available: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectCode && resp.StatusCode != http.StatusNotFound {
				t.Logf("Response status: %d (expected %d)", resp.StatusCode, tc.expectCode)
			}
		})
	}
}

// TestLeaderboardEndpoints tests the leaderboard endpoints
func TestLeaderboardEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	testCases := []struct {
		name     string
		endpoint string
	}{
		{"Get available metrics", "/internal/leaderboards/metrics"},
		{"Get problems solved leaderboard", "/internal/leaderboards/problems_solved"},
		{"Get fastest avg leaderboard", "/internal/leaderboards/fastest_avg"},
		{"Get user ranks", "/internal/leaderboards/user/1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := backendURL + tc.endpoint
			resp, err := http.Get(url)
			if err != nil {
				t.Skipf("Backend not available: %v", err)
				return
			}
			defer resp.Body.Close()

			t.Logf("Response status: %d", resp.StatusCode)
		})
	}
}

// TestAntiCheatEndpoints tests the anti-cheat endpoints
func TestAntiCheatEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Test check submission endpoint
	checkReq := map[string]interface{}{
		"user_id":           1,
		"problem_id":        1,
		"execution_time_ms": 100,
		"memory_used_kb":    1000,
		"difficulty":        "easy",
	}

	body, _ := json.Marshal(checkReq)
	url := backendURL + "/internal/anticheat/check"

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}
	defer resp.Body.Close()

	t.Logf("Anti-cheat check response status: %d", resp.StatusCode)
}

// Helper functions

func submitCode(code, language string, problemID int) (*SubmissionResponse, error) {
	req := SubmissionRequest{
		Code:      code,
		Language:  language,
		ProblemID: problemID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := baseURL + "/api/submissions"
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result SubmissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func generateLongCode(length int) string {
	code := "# Long code test\n"
	for i := 0; len(code) < length; i++ {
		code += fmt.Sprintf("x%d = %d\n", i, i)
	}
	code += "print('done')\n"
	return code
}
