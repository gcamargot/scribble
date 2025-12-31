package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ExecutionJobParams contains parameters for creating an execution job
type ExecutionJobParams struct {
	SubmissionID string      // Unique submission ID
	ProblemID    string      // Problem identifier
	Code         string      // User's code to execute
	TestCases    interface{} // Test cases to run against
}

// JobManager handles Kubernetes Job operations
type JobManager struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewJobManager creates a new K8s Job manager
// Automatically detects in-cluster config or falls back to kubeconfig
func NewJobManager() (*JobManager, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first (for production K8s deployment)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig (for local development)
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build K8s config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s clientset: %w", err)
	}

	namespace := os.Getenv("K8S_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	return &JobManager{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// CreateExecutionJob creates a Kubernetes Job to execute user code
// Job name format: exec-<submission-id>-<random>
// Resource limits: 512Mi memory, 500m CPU
// ActiveDeadlineSeconds: 10
func (jm *JobManager) CreateExecutionJob(ctx context.Context, params ExecutionJobParams) (string, error) {
	// Generate random suffix for job name uniqueness
	rand.Seed(time.Now().UnixNano())
	randomSuffix := fmt.Sprintf("%06d", rand.Intn(1000000))
	jobName := fmt.Sprintf("exec-%s-%s", params.SubmissionID, randomSuffix)

	// Encode code to base64 for safe env var transmission
	codeB64 := base64.StdEncoding.EncodeToString([]byte(params.Code))

	// Marshal test cases to JSON
	testCasesJSON, err := json.Marshal(params.TestCases)
	if err != nil {
		return "", fmt.Errorf("failed to marshal test cases: %w", err)
	}

	// Get executor image from environment or use default
	executorImage := os.Getenv("PYTHON_EXECUTOR_IMAGE")
	if executorImage == "" {
		executorImage = "scribble-python-executor:latest"
	}

	// Define resource limits
	cpuLimit := resource.MustParse("500m")
	memoryLimit := resource.MustParse("512Mi")

	// ActiveDeadlineSeconds: kill job after 10 seconds
	var activeDeadlineSeconds int64 = 10

	// Get RuntimeClass from environment (gVisor for sandboxed execution)
	runtimeClassName := os.Getenv("K8S_RUNTIME_CLASS")
	if runtimeClassName == "" {
		runtimeClassName = "gvisor" // Default to gVisor for security
	}

	// Security context for container isolation
	// - Run as non-root user (UID 1000 matches coderunner in executor image)
	// - Read-only root filesystem prevents malicious writes
	// - No privilege escalation allowed
	runAsNonRoot := true
	runAsUser := int64(1000)
	readOnlyRootFilesystem := true
	allowPrivilegeEscalation := false

	// Job specification
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: jm.namespace,
			Labels: map[string]string{
				"app":           "scribble-executor",
				"submission-id": params.SubmissionID,
				"problem-id":    params.ProblemID,
			},
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds: &activeDeadlineSeconds,
			// Don't retry failed executions automatically
			BackoffLimit: func() *int32 { i := int32(0); return &i }(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":           "scribble-executor",
						"submission-id": params.SubmissionID,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyNever,
					RuntimeClassName: &runtimeClassName,
					// Disable automounting service account token (not needed for code execution)
					AutomountServiceAccountToken: func() *bool { b := false; return &b }(),
					// Pod-level security context
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &runAsNonRoot,
						RunAsUser:    &runAsUser,
						// Ensure all containers run with same security restrictions
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					// Disable DNS resolution (containers don't need network)
					DNSPolicy: corev1.DNSNone,
					DNSConfig: &corev1.PodDNSConfig{
						Nameservers: []string{},
					},
					Containers: []corev1.Container{
						{
							Name:  "executor",
							Image: executorImage,
							Env: []corev1.EnvVar{
								{
									Name:  "CODE",
									Value: codeB64,
								},
								{
									Name:  "TEST_CASES",
									Value: string(testCasesJSON),
								},
								{
									Name:  "PROBLEM_ID",
									Value: params.ProblemID,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memoryLimit,
								},
								Requests: corev1.ResourceList{
									// Request same as limit for guaranteed QoS
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memoryLimit,
								},
							},
							// Container-level security context
							SecurityContext: &corev1.SecurityContext{
								ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
								RunAsNonRoot:             &runAsNonRoot,
								RunAsUser:                &runAsUser,
								Capabilities: &corev1.Capabilities{
									// Drop all capabilities for maximum isolation
									Drop: []corev1.Capability{"ALL"},
								},
							},
							// Write /tmp for temporary files (required by some languages)
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "tmp",
									MountPath: "/tmp",
								},
							},
						},
					},
					// EmptyDir volume for /tmp (ephemeral, secure)
					Volumes: []corev1.Volume{
						{
							Name: "tmp",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									// Limit /tmp to 64Mi to prevent disk abuse
									SizeLimit: func() *resource.Quantity {
										q := resource.MustParse("64Mi")
										return &q
									}(),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the job in Kubernetes
	createdJob, err := jm.clientset.BatchV1().Jobs(jm.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create K8s job: %w", err)
	}

	return createdJob.Name, nil
}

// GetJobLogs retrieves logs from a completed job
func (jm *JobManager) GetJobLogs(ctx context.Context, jobName string) (string, error) {
	// Get pods for this job
	pods, err := jm.clientset.CoreV1().Pods(jm.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list pods for job: %w", err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for job %s", jobName)
	}

	// Get logs from the first pod (jobs should only have one pod)
	podName := pods.Items[0].Name
	req := jm.clientset.CoreV1().Pods(jm.namespace).GetLogs(podName, &corev1.PodLogOptions{})
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}
	defer logs.Close()

	// Read logs
	buf := make([]byte, 4096)
	n, err := logs.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return string(buf[:n]), nil
}

// DeleteJob removes a job and its associated pods
func (jm *JobManager) DeleteJob(ctx context.Context, jobName string) error {
	// Delete with PropagationPolicy=Background to clean up pods automatically
	deletePolicy := metav1.DeletePropagationBackground
	err := jm.clientset.BatchV1().Jobs(jm.namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}
	return nil
}
