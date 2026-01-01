package load

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	defaultBackendURL = "http://localhost:8080"
	defaultWorkers    = 10
	defaultRequests   = 100
)

var backendURL string

func init() {
	backendURL = os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = defaultBackendURL
	}
}

// LoadTestResult tracks performance metrics for a load test
type LoadTestResult struct {
	TotalRequests   int64
	SuccessCount    int64
	ErrorCount      int64
	Latencies       []time.Duration
	StartTime       time.Time
	EndTime         time.Time
	mu              sync.Mutex
}

// NewLoadTestResult creates a new result tracker
func NewLoadTestResult() *LoadTestResult {
	return &LoadTestResult{
		Latencies: make([]time.Duration, 0, defaultRequests),
	}
}

// RecordSuccess records a successful request
func (r *LoadTestResult) RecordSuccess(latency time.Duration) {
	atomic.AddInt64(&r.TotalRequests, 1)
	atomic.AddInt64(&r.SuccessCount, 1)
	r.mu.Lock()
	r.Latencies = append(r.Latencies, latency)
	r.mu.Unlock()
}

// RecordError records a failed request
func (r *LoadTestResult) RecordError() {
	atomic.AddInt64(&r.TotalRequests, 1)
	atomic.AddInt64(&r.ErrorCount, 1)
}

// GetStats calculates statistics from the load test
func (r *LoadTestResult) GetStats() map[string]interface{} {
	r.mu.Lock()
	latencies := make([]time.Duration, len(r.Latencies))
	copy(latencies, r.Latencies)
	r.mu.Unlock()

	if len(latencies) == 0 {
		return map[string]interface{}{
			"total_requests": r.TotalRequests,
			"success_count":  r.SuccessCount,
			"error_count":    r.ErrorCount,
			"error_rate":     float64(r.ErrorCount) / float64(r.TotalRequests) * 100,
		}
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	duration := r.EndTime.Sub(r.StartTime)
	rps := float64(r.TotalRequests) / duration.Seconds()

	var totalLatency time.Duration
	for _, l := range latencies {
		totalLatency += l
	}
	avgLatency := totalLatency / time.Duration(len(latencies))

	return map[string]interface{}{
		"total_requests": r.TotalRequests,
		"success_count":  r.SuccessCount,
		"error_count":    r.ErrorCount,
		"error_rate":     fmt.Sprintf("%.2f%%", float64(r.ErrorCount)/float64(r.TotalRequests)*100),
		"duration":       duration.String(),
		"requests_per_s": fmt.Sprintf("%.2f", rps),
		"latency_min":    latencies[0].String(),
		"latency_max":    latencies[len(latencies)-1].String(),
		"latency_avg":    avgLatency.String(),
		"latency_p50":    latencies[len(latencies)/2].String(),
		"latency_p95":    latencies[int(float64(len(latencies))*0.95)].String(),
		"latency_p99":    latencies[int(float64(len(latencies))*0.99)].String(),
	}
}

// PrintStats outputs formatted statistics
func (r *LoadTestResult) PrintStats(t *testing.T) {
	stats := r.GetStats()
	t.Logf("=== Load Test Results ===")
	t.Logf("Total Requests: %d", stats["total_requests"])
	t.Logf("Successful: %d", stats["success_count"])
	t.Logf("Errors: %d", stats["error_count"])
	t.Logf("Error Rate: %s", stats["error_rate"])
	t.Logf("Duration: %s", stats["duration"])
	t.Logf("Requests/sec: %s", stats["requests_per_s"])
	if _, ok := stats["latency_avg"]; ok {
		t.Logf("Latency (min/avg/max): %s / %s / %s",
			stats["latency_min"], stats["latency_avg"], stats["latency_max"])
		t.Logf("Latency (p50/p95/p99): %s / %s / %s",
			stats["latency_p50"], stats["latency_p95"], stats["latency_p99"])
	}
}

// makeRequest makes an HTTP request and records the result
func makeRequest(url string, result *LoadTestResult) {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		result.RecordError()
		return
	}
	defer resp.Body.Close()

	// Read body to properly measure full response time
	_, _ = io.ReadAll(resp.Body)

	latency := time.Since(start)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.RecordSuccess(latency)
	} else {
		result.RecordError()
	}
}

// runLoadTest executes a load test with the given configuration
func runLoadTest(t *testing.T, url string, numRequests, numWorkers int) *LoadTestResult {
	result := NewLoadTestResult()
	result.StartTime = time.Now()

	var wg sync.WaitGroup
	requestChan := make(chan struct{}, numRequests)

	// Spawn workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestChan {
				makeRequest(url, result)
			}
		}()
	}

	// Send requests
	for i := 0; i < numRequests; i++ {
		requestChan <- struct{}{}
	}
	close(requestChan)

	wg.Wait()
	result.EndTime = time.Now()

	return result
}

// TestLeaderboardLoad tests leaderboard endpoints under load
func TestLeaderboardLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	// Test problems_solved leaderboard (most common query)
	url := fmt.Sprintf("%s/api/leaderboard?metric=problems_solved&page=1&page_size=50", backendURL)

	t.Logf("Testing leaderboard endpoint: %s", url)
	t.Logf("Config: 1000 requests, 50 workers (simulating 1000 users)")

	result := runLoadTest(t, url, 1000, 50)
	result.PrintStats(t)

	// Assert performance targets
	stats := result.GetStats()
	if result.ErrorCount > int64(float64(result.TotalRequests)*0.01) {
		t.Errorf("Error rate too high: %s (target: <1%%)", stats["error_rate"])
	}
}

// TestLeaderboardLoadAllMetrics tests all leaderboard metric types
func TestLeaderboardLoadAllMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	metrics := []string{"problems_solved", "fastest_avg", "lowest_memory_avg", "longest_streak"}

	for _, metric := range metrics {
		t.Run(metric, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/leaderboard?metric=%s&page=1&page_size=20", backendURL, metric)

			result := runLoadTest(t, url, 100, 10)
			result.PrintStats(t)

			if result.ErrorCount > 5 {
				t.Errorf("Too many errors for metric %s: %d", metric, result.ErrorCount)
			}
		})
	}
}

// TestConcurrentPagination tests concurrent access to different pages
func TestConcurrentPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	result := NewLoadTestResult()
	result.StartTime = time.Now()

	var wg sync.WaitGroup
	numPages := 50
	requestsPerPage := 20

	for page := 1; page <= numPages; page++ {
		for i := 0; i < requestsPerPage; i++ {
			wg.Add(1)
			go func(p int) {
				defer wg.Done()
				url := fmt.Sprintf("%s/api/leaderboard?metric=problems_solved&page=%d&page_size=20", backendURL, p)
				makeRequest(url, result)
			}(page)
		}
	}

	wg.Wait()
	result.EndTime = time.Now()

	t.Logf("Testing concurrent pagination (50 pages x 20 requests)")
	result.PrintStats(t)
}

// TestUserRankLoad tests user rank lookups under load
func TestUserRankLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	result := NewLoadTestResult()
	result.StartTime = time.Now()

	var wg sync.WaitGroup

	// Simulate 100 different users checking their ranks
	for userID := 1; userID <= 100; userID++ {
		for metric := 0; metric < 4; metric++ {
			wg.Add(1)
			go func(uid, m int) {
				defer wg.Done()
				metrics := []string{"problems_solved", "fastest_avg", "lowest_memory_avg", "longest_streak"}
				url := fmt.Sprintf("%s/api/leaderboard/user/%d?metric=%s", backendURL, uid, metrics[m])
				makeRequest(url, result)
			}(userID, metric)
		}
	}

	wg.Wait()
	result.EndTime = time.Now()

	t.Logf("Testing user rank lookups (100 users x 4 metrics = 400 requests)")
	result.PrintStats(t)
}

// TestHealthCheck tests the health endpoint under load
func TestHealthCheck(t *testing.T) {
	url := fmt.Sprintf("%s/health", backendURL)

	result := runLoadTest(t, url, 500, 50)
	result.PrintStats(t)

	if result.ErrorCount > 0 {
		t.Errorf("Health check should never fail, got %d errors", result.ErrorCount)
	}
}

// BenchmarkLeaderboardQuery benchmarks leaderboard query performance
func BenchmarkLeaderboardQuery(b *testing.B) {
	url := fmt.Sprintf("%s/api/leaderboard?metric=problems_solved&page=1&page_size=50", backendURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(url)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkUserRankLookup benchmarks user rank lookup performance
func BenchmarkUserRankLookup(b *testing.B) {
	url := fmt.Sprintf("%s/api/leaderboard/user/1?metric=problems_solved", backendURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(url)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}

// TestStressTest runs an aggressive stress test
func TestStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Only run if explicitly enabled
	if os.Getenv("RUN_STRESS_TEST") != "true" {
		t.Skip("Stress test disabled. Set RUN_STRESS_TEST=true to enable.")
	}

	t.Log("Running stress test: 10000 requests with 100 workers")
	url := fmt.Sprintf("%s/api/leaderboard?metric=problems_solved&page=1", backendURL)

	result := runLoadTest(t, url, 10000, 100)
	result.PrintStats(t)

	// In stress test, we're more lenient with error rate
	if result.ErrorCount > int64(float64(result.TotalRequests)*0.05) {
		t.Errorf("Error rate too high under stress: got %d errors", result.ErrorCount)
	}
}

// TestResponseValidation validates response structure under load
func TestResponseValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping validation test in short mode")
	}

	url := fmt.Sprintf("%s/api/leaderboard?metric=problems_solved&page=1&page_size=10", backendURL)

	var validationErrors int64
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := http.Get(url)
			if err != nil {
				mu.Lock()
				validationErrors++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				mu.Lock()
				validationErrors++
				mu.Unlock()
				return
			}

			// Validate JSON structure
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				mu.Lock()
				validationErrors++
				mu.Unlock()
				return
			}

			// Check required fields exist
			requiredFields := []string{"entries", "metric_type", "page", "total"}
			for _, field := range requiredFields {
				if _, exists := result[field]; !exists {
					mu.Lock()
					validationErrors++
					mu.Unlock()
					return
				}
			}
		}()
	}

	wg.Wait()

	if validationErrors > 0 {
		t.Errorf("Got %d validation errors in 50 requests", validationErrors)
	}
}
