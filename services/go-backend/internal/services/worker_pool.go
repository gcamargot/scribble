package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nahtao97/scribble/internal/k8s"
)

// Worker pool configuration constants
const (
	DefaultWorkerCount = 10
	DefaultQueueSize   = 100
	DefaultRateLimit   = 5  // submissions per minute per user
	DefaultRateWindow  = time.Minute
)

// Common errors
var (
	ErrQueueFull     = errors.New("submission queue is full")
	ErrRateLimited   = errors.New("rate limit exceeded")
	ErrPoolShutdown  = errors.New("worker pool is shutting down")
)

// ExecutionJob represents a job in the worker queue
type ExecutionJob struct {
	Params     k8s.ExecutionJobParams
	ResultChan chan *ExecutionJobResult
	Ctx        context.Context
}

// ExecutionJobResult is the result of an execution job
type ExecutionJobResult struct {
	Result *k8s.ExecutionResult
	Error  error
}

// RateLimiter tracks request rates per user
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request from userID is allowed
func (r *RateLimiter) Allow(userID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	// Get existing requests and filter out old ones
	reqs := r.requests[userID]
	validReqs := make([]time.Time, 0, len(reqs))
	for _, t := range reqs {
		if t.After(windowStart) {
			validReqs = append(validReqs, t)
		}
	}

	// Check if under limit
	if len(validReqs) >= r.limit {
		r.requests[userID] = validReqs
		return false
	}

	// Add current request
	validReqs = append(validReqs, now)
	r.requests[userID] = validReqs
	return true
}

// Reset clears the rate limiter (useful for testing)
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = make(map[string][]time.Time)
}

// WorkerPool manages concurrent code execution with rate limiting
type WorkerPool struct {
	jobManager  *k8s.JobManager
	jobQueue    chan *ExecutionJob
	rateLimiter *RateLimiter
	workerCount int
	wg          sync.WaitGroup
	shutdown    chan struct{}
	isShutdown  bool
	mu          sync.RWMutex
}

// WorkerPoolConfig configures the worker pool
type WorkerPoolConfig struct {
	WorkerCount int
	QueueSize   int
	RateLimit   int           // requests per window
	RateWindow  time.Duration // rate limit window
}

// DefaultWorkerPoolConfig returns default configuration
func DefaultWorkerPoolConfig() WorkerPoolConfig {
	return WorkerPoolConfig{
		WorkerCount: DefaultWorkerCount,
		QueueSize:   DefaultQueueSize,
		RateLimit:   DefaultRateLimit,
		RateWindow:  DefaultRateWindow,
	}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(jobManager *k8s.JobManager, config WorkerPoolConfig) *WorkerPool {
	wp := &WorkerPool{
		jobManager:  jobManager,
		jobQueue:    make(chan *ExecutionJob, config.QueueSize),
		rateLimiter: NewRateLimiter(config.RateLimit, config.RateWindow),
		workerCount: config.WorkerCount,
		shutdown:    make(chan struct{}),
	}

	// Start workers
	for i := 0; i < config.WorkerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	return wp
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.shutdown:
			return
		case job, ok := <-wp.jobQueue:
			if !ok {
				return
			}
			wp.processJob(job)
		}
	}
}

// processJob executes a single job
func (wp *WorkerPool) processJob(job *ExecutionJob) {
	result, err := wp.jobManager.ExecuteAndWait(job.Ctx, job.Params)
	job.ResultChan <- &ExecutionJobResult{
		Result: result,
		Error:  err,
	}
	close(job.ResultChan)
}

// Submit submits a job to the worker pool
// Returns ErrQueueFull if the queue is at capacity (circuit breaker)
// Returns ErrRateLimited if the user has exceeded their rate limit
func (wp *WorkerPool) Submit(ctx context.Context, userID string, params k8s.ExecutionJobParams) (*k8s.ExecutionResult, error) {
	wp.mu.RLock()
	if wp.isShutdown {
		wp.mu.RUnlock()
		return nil, ErrPoolShutdown
	}
	wp.mu.RUnlock()

	// Check rate limit
	if !wp.rateLimiter.Allow(userID) {
		return nil, ErrRateLimited
	}

	// Create result channel
	resultChan := make(chan *ExecutionJobResult, 1)

	// Create job
	job := &ExecutionJob{
		Params:     params,
		ResultChan: resultChan,
		Ctx:        ctx,
	}

	// Try to add to queue (non-blocking)
	select {
	case wp.jobQueue <- job:
		// Job queued successfully
	default:
		// Queue is full - circuit breaker trips
		return nil, ErrQueueFull
	}

	// Wait for result
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultChan:
		return result.Result, result.Error
	}
}

// QueueLength returns the current number of jobs in the queue
func (wp *WorkerPool) QueueLength() int {
	return len(wp.jobQueue)
}

// QueueCapacity returns the maximum queue capacity
func (wp *WorkerPool) QueueCapacity() int {
	return cap(wp.jobQueue)
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown(ctx context.Context) error {
	wp.mu.Lock()
	if wp.isShutdown {
		wp.mu.Unlock()
		return nil
	}
	wp.isShutdown = true
	wp.mu.Unlock()

	// Signal workers to stop
	close(wp.shutdown)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stats returns current worker pool statistics
type PoolStats struct {
	WorkerCount   int `json:"worker_count"`
	QueueLength   int `json:"queue_length"`
	QueueCapacity int `json:"queue_capacity"`
	IsShutdown    bool `json:"is_shutdown"`
}

// Stats returns current pool statistics
func (wp *WorkerPool) Stats() PoolStats {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	return PoolStats{
		WorkerCount:   wp.workerCount,
		QueueLength:   len(wp.jobQueue),
		QueueCapacity: cap(wp.jobQueue),
		IsShutdown:    wp.isShutdown,
	}
}
