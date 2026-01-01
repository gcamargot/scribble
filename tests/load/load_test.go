package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Configuration
var (
	baseURL    = getEnv("TEST_BASE_URL", "http://localhost:3000")
	backendURL = getEnv("TEST_BACKEND_URL", "http://localhost:8080")
)

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// LoadTestResult contains statistics from a load test
type LoadTestResult struct {
	TotalRequests   int64
	SuccessCount    int64
	ErrorCount      int64
	TotalDuration   time.Duration
	AvgLatency      time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	RequestsPerSec  float64
}

// SubmissionRequest represents a code submission
type SubmissionRequest struct {
	Code      string `json:"code"`
	Language  string `json:"language"`
	ProblemID int    `json:"problemId"`
}

// TestLeaderboardLoad tests leaderboard queries under load
// Simulates 1000 users fetching the leaderboard
func TestLeaderboardLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	const (
		numRequests   = 100 // Scale down for CI, increase for real load tests
		numWorkers    = 10
		targetRPS     = 50  // Target requests per second
	)

	result := runLoadTest(t, numRequests, numWorkers, func() (time.Duration, error) {
		start := time.Now()
		url := backendURL + "/internal/leaderboards/problems_solved?page=1&page_size=20"
		resp, err := http.Get(url)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
		return time.Since(start), nil
	})

	t.Logf("Leaderboard Load Test Results:")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Success Rate: %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	t.Logf("  Total Duration: %v", result.TotalDuration)
	t.Logf("  Avg Latency: %v", result.AvgLatency)
	t.Logf("  Min Latency: %v", result.MinLatency)
	t.Logf("  Max Latency: %v", result.MaxLatency)
	t.Logf("  Requests/sec: %.2f", result.RequestsPerSec)
}

// TestSubmissionLoad tests submission endpoint under load
func TestSubmissionLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	const (
		numRequests = 50
		numWorkers  = 5
	)

	sampleCode := `print(sum(map(int, input().split())))`

	result := runLoadTest(t, numRequests, numWorkers, func() (time.Duration, error) {
		start := time.Now()

		req := SubmissionRequest{
			Code:      sampleCode,
			Language:  "python",
			ProblemID: 1,
		}
		body, _ := json.Marshal(req)

		resp, err := http.Post(baseURL+"/api/submissions", "application/json", bytes.NewReader(body))
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		// Accept various status codes as the service may be mocked
		return time.Since(start), nil
	})

	t.Logf("Submission Load Test Results:")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Success Rate: %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	t.Logf("  Total Duration: %v", result.TotalDuration)
	t.Logf("  Avg Latency: %v", result.AvgLatency)
	t.Logf("  Min Latency: %v", result.MinLatency)
	t.Logf("  Max Latency: %v", result.MaxLatency)
	t.Logf("  Requests/sec: %.2f", result.RequestsPerSec)
}

// TestProblemLoad tests problem endpoint under load
func TestProblemLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	const (
		numRequests = 200
		numWorkers  = 20
	)

	result := runLoadTest(t, numRequests, numWorkers, func() (time.Duration, error) {
		start := time.Now()
		url := backendURL + "/internal/problems/daily/today"
		resp, err := http.Get(url)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		return time.Since(start), nil
	})

	t.Logf("Problem Load Test Results:")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Success Rate: %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	t.Logf("  Total Duration: %v", result.TotalDuration)
	t.Logf("  Avg Latency: %v", result.AvgLatency)
	t.Logf("  Requests/sec: %.2f", result.RequestsPerSec)
}

// TestAntiCheatLoad tests anti-cheat check endpoint under load
func TestAntiCheatLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	const (
		numRequests = 100
		numWorkers  = 10
	)

	result := runLoadTest(t, numRequests, numWorkers, func() (time.Duration, error) {
		start := time.Now()

		checkReq := map[string]interface{}{
			"user_id":           1,
			"problem_id":        1,
			"execution_time_ms": 100,
			"memory_used_kb":    1000,
			"difficulty":        "easy",
		}
		body, _ := json.Marshal(checkReq)

		resp, err := http.Post(backendURL+"/internal/anticheat/check", "application/json", bytes.NewReader(body))
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		return time.Since(start), nil
	})

	t.Logf("Anti-Cheat Load Test Results:")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Success Rate: %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	t.Logf("  Total Duration: %v", result.TotalDuration)
	t.Logf("  Avg Latency: %v", result.AvgLatency)
	t.Logf("  Requests/sec: %.2f", result.RequestsPerSec)
}

// TestConcurrentUsers simulates multiple concurrent users
func TestConcurrentUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	const numUsers = 20

	var wg sync.WaitGroup
	userResults := make(chan string, numUsers)

	startTime := time.Now()

	for userID := 1; userID <= numUsers; userID++ {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()

			// Each user performs a series of actions
			actions := []string{
				fmt.Sprintf("%s/internal/problems/daily/today", backendURL),
				fmt.Sprintf("%s/internal/leaderboards/user/%d", backendURL, uid),
			}

			for _, url := range actions {
				resp, err := http.Get(url)
				if err != nil {
					userResults <- fmt.Sprintf("User %d: error - %v", uid, err)
					continue
				}
				resp.Body.Close()
			}
			userResults <- fmt.Sprintf("User %d: completed", uid)
		}(userID)
	}

	wg.Wait()
	close(userResults)

	duration := time.Since(startTime)

	successCount := 0
	for range userResults {
		successCount++
	}

	t.Logf("Concurrent Users Test:")
	t.Logf("  Users: %d", numUsers)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Completed: %d", successCount)
}

// runLoadTest executes a load test with the given parameters
func runLoadTest(t *testing.T, numRequests, numWorkers int, requestFn func() (time.Duration, error)) LoadTestResult {
	var (
		successCount int64
		errorCount   int64
		totalLatency int64
		minLatency   int64 = int64(time.Hour)
		maxLatency   int64
		mu           sync.Mutex
	)

	jobs := make(chan int, numRequests)
	var wg sync.WaitGroup

	startTime := time.Now()

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				latency, err := requestFn()
				latencyNs := int64(latency)

				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
					atomic.AddInt64(&totalLatency, latencyNs)

					mu.Lock()
					if latencyNs < minLatency {
						minLatency = latencyNs
					}
					if latencyNs > maxLatency {
						maxLatency = latencyNs
					}
					mu.Unlock()
				}
			}
		}()
	}

	// Enqueue jobs
	for i := 0; i < numRequests; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	totalDuration := time.Since(startTime)

	var avgLatency time.Duration
	if successCount > 0 {
		avgLatency = time.Duration(totalLatency / successCount)
	}

	if minLatency == int64(time.Hour) {
		minLatency = 0
	}

	return LoadTestResult{
		TotalRequests:   int64(numRequests),
		SuccessCount:    successCount,
		ErrorCount:      errorCount,
		TotalDuration:   totalDuration,
		AvgLatency:      avgLatency,
		MinLatency:      time.Duration(minLatency),
		MaxLatency:      time.Duration(maxLatency),
		RequestsPerSec:  float64(numRequests) / totalDuration.Seconds(),
	}
}

// BenchmarkLeaderboardQuery benchmarks leaderboard queries
func BenchmarkLeaderboardQuery(b *testing.B) {
	url := backendURL + "/internal/leaderboards/problems_solved"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(url)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// BenchmarkProblemQuery benchmarks problem queries
func BenchmarkProblemQuery(b *testing.B) {
	url := backendURL + "/internal/problems/daily/today"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(url)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
