package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// ExecutionResult represents the result of code execution
type ExecutionResult struct {
	Status          string       `json:"status"`           // accepted, wrong_answer, runtime_error, time_limit, memory_limit, compilation_error
	ErrorMessage    string       `json:"error_message"`    // Error details if any
	ExecutionTimeMs int64        `json:"execution_time_ms"`
	MemoryUsedKB    int64        `json:"memory_used_kb"`
	TestsPassed     int          `json:"tests_passed"`
	TestsTotal      int          `json:"tests_total"`
	TestResults     []TestResult `json:"test_results,omitempty"`
}

// TestResult represents the result of a single test case
type TestResult struct {
	TestCaseID   int         `json:"test_case_id"`
	Passed       bool        `json:"passed"`
	ActualOutput interface{} `json:"actual_output,omitempty"`
	Error        string      `json:"error,omitempty"`
}

// Common error types for job monitoring
var (
	ErrJobTimeout     = errors.New("job exceeded time limit")
	ErrJobFailed      = errors.New("job execution failed")
	ErrPodCrashed     = errors.New("executor pod crashed")
	ErrNoLogsFound    = errors.New("no execution logs found")
	ErrInvalidResult  = errors.New("failed to parse execution result")
	ErrJobNotFound    = errors.New("job not found")
)

// MonitorConfig configures job monitoring behavior
type MonitorConfig struct {
	// MaxWaitTime is the maximum time to wait for job completion
	MaxWaitTime time.Duration
	// PollInterval is how often to check job status (fallback if watch fails)
	PollInterval time.Duration
}

// DefaultMonitorConfig returns sensible defaults for monitoring
func DefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		MaxWaitTime:  15 * time.Second,
		PollInterval: 500 * time.Millisecond,
	}
}

// WaitForJobCompletion waits for a job to complete and returns the execution result
// It uses the Kubernetes watch API for efficient notification of job status changes
func (jm *JobManager) WaitForJobCompletion(ctx context.Context, jobName string, config MonitorConfig) (*ExecutionResult, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, config.MaxWaitTime)
	defer cancel()

	// Set up watch on the specific job
	watcher, err := jm.clientset.BatchV1().Jobs(jm.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", jobName),
	})
	if err != nil {
		// Fall back to polling if watch fails
		return jm.pollForJobCompletion(ctx, jobName, config)
	}
	defer watcher.Stop()

	// Watch for job status changes
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return &ExecutionResult{
					Status:       "time_limit",
					ErrorMessage: "Execution exceeded time limit",
				}, ErrJobTimeout
			}
			return nil, ctx.Err()

		case event, ok := <-watcher.ResultChan():
			if !ok {
				// Watcher closed, fall back to polling
				return jm.pollForJobCompletion(ctx, jobName, config)
			}

			if event.Type == watch.Error {
				continue
			}

			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				continue
			}

			// Check if job completed
			result, done, err := jm.checkJobStatus(ctx, job)
			if done {
				return result, err
			}
		}
	}
}

// pollForJobCompletion polls job status at regular intervals
func (jm *JobManager) pollForJobCompletion(ctx context.Context, jobName string, config MonitorConfig) (*ExecutionResult, error) {
	ticker := time.NewTicker(config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return &ExecutionResult{
					Status:       "time_limit",
					ErrorMessage: "Execution exceeded time limit",
				}, ErrJobTimeout
			}
			return nil, ctx.Err()

		case <-ticker.C:
			job, err := jm.clientset.BatchV1().Jobs(jm.namespace).Get(ctx, jobName, metav1.GetOptions{})
			if err != nil {
				continue
			}

			result, done, err := jm.checkJobStatus(ctx, job)
			if done {
				return result, err
			}
		}
	}
}

// checkJobStatus examines job status and returns result if complete
func (jm *JobManager) checkJobStatus(ctx context.Context, job *batchv1.Job) (*ExecutionResult, bool, error) {
	// Check for successful completion
	if job.Status.Succeeded > 0 {
		result, err := jm.collectJobResult(ctx, job.Name)
		return result, true, err
	}

	// Check for failure
	if job.Status.Failed > 0 {
		// Try to get error details from pod
		result, err := jm.collectJobResult(ctx, job.Name)
		if err != nil {
			// If we can't get logs, return a generic failure
			return &ExecutionResult{
				Status:       "runtime_error",
				ErrorMessage: "Execution failed",
			}, true, ErrJobFailed
		}
		return result, true, nil
	}

	// Check for active deadline exceeded
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Reason == "DeadlineExceeded" {
			return &ExecutionResult{
				Status:       "time_limit",
				ErrorMessage: "Execution exceeded time limit",
			}, true, ErrJobTimeout
		}
	}

	// Job still running
	return nil, false, nil
}

// collectJobResult reads the execution result from pod logs
func (jm *JobManager) collectJobResult(ctx context.Context, jobName string) (*ExecutionResult, error) {
	// Find the pod for this job
	pods, err := jm.clientset.CoreV1().Pods(jm.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, ErrNoLogsFound
	}

	pod := &pods.Items[0]

	// Check if pod crashed
	if pod.Status.Phase == corev1.PodFailed {
		// Check for OOMKilled
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Terminated != nil {
				if containerStatus.State.Terminated.Reason == "OOMKilled" {
					return &ExecutionResult{
						Status:       "memory_limit",
						ErrorMessage: "Execution exceeded memory limit",
					}, nil
				}
			}
		}
	}

	// Get pod logs
	logs, err := jm.getPodLogs(ctx, pod.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}

	if logs == "" {
		return nil, ErrNoLogsFound
	}

	// Parse execution result JSON from logs
	var result ExecutionResult
	if err := json.Unmarshal([]byte(logs), &result); err != nil {
		// If we can't parse as JSON, treat as runtime error
		return &ExecutionResult{
			Status:       "runtime_error",
			ErrorMessage: fmt.Sprintf("Failed to parse execution result: %s", truncateString(logs, 200)),
		}, nil
	}

	return &result, nil
}

// getPodLogs retrieves complete logs from a pod
func (jm *JobManager) getPodLogs(ctx context.Context, podName string) (string, error) {
	req := jm.clientset.CoreV1().Pods(jm.namespace).GetLogs(podName, &corev1.PodLogOptions{})
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	logs, err := io.ReadAll(stream)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// CleanupJob deletes a completed job and its pods
func (jm *JobManager) CleanupJob(ctx context.Context, jobName string) error {
	propagation := metav1.DeletePropagationBackground
	return jm.clientset.BatchV1().Jobs(jm.namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
}

// ExecuteAndWait is a convenience method that creates a job, waits for completion,
// collects the result, and cleans up
func (jm *JobManager) ExecuteAndWait(ctx context.Context, params ExecutionJobParams) (*ExecutionResult, error) {
	config := DefaultMonitorConfig()

	// Create the job
	jobName, err := jm.CreateExecutionJob(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Always cleanup the job when done
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = jm.CleanupJob(cleanupCtx, jobName)
	}()

	// Wait for completion and collect result
	result, err := jm.WaitForJobCompletion(ctx, jobName, config)
	if err != nil && result == nil {
		return &ExecutionResult{
			Status:       "runtime_error",
			ErrorMessage: err.Error(),
		}, err
	}

	return result, nil
}
