package jobs

import (
	"time"
)

// Framework represents the ML training framework
type Framework string

const (
	FrameworkPyTorch    Framework = "pytorch"
	FrameworkTensorFlow Framework = "tensorflow"
	FrameworkXGBoost    Framework = "xgboost"
	FrameworkMPI        Framework = "mpi" // Horovod, DeepSpeed
)

// Status represents job status
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// StorageConfig represents storage configuration
type StorageConfig struct {
	DatasetS3Path    string `json:"dataset_s3_path"`
	CheckpointS3Path string `json:"checkpoint_s3_path"`
	FSxCapacityGB    int    `json:"fsx_capacity_gb"`
}

// Job represents a distributed training job
type Job struct {
	ID            string        `json:"id"`
	ProjectID     string        `json:"project_id"`
	Framework     Framework     `json:"framework"`
	Image         string        `json:"image"`
	Command       []string      `json:"command,omitempty"`
	Args          []string      `json:"args,omitempty"`
	Workers       int           `json:"workers"`
	GPUsPerWorker int           `json:"gpus_per_worker"`
	EFAEnabled    bool          `json:"efa_enabled"`
	Storage       StorageConfig `json:"storage"`
	Status        Status        `json:"status"`
	StatusMessage string        `json:"status_message,omitempty"`
	CostUSD       float64       `json:"cost_usd,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	StartedAt     *time.Time    `json:"started_at,omitempty"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
}

// Repository interface for job storage
type Repository interface {
	Create(job *Job) (*Job, error)
	Get(id string) (*Job, error)
	List(projectID string, status Status) ([]*Job, error)
	UpdateStatus(id string, status Status, message string) error
	UpdateCost(id string, costUSD float64) error
	Delete(id string) error
}

// Queue interface for job queue
type Queue interface {
	Push(job *Job) error
}

// Executor interface for job execution
// Uses Kubeflow Training Operator CRDs under the hood
type Executor interface {
	Execute(job *Job) error
	Cancel(jobID string) error
	GetLogs(jobID string) (string, error)
}

// KubeflowTemplates maps framework to Kubeflow CRD template
// Templates are in templates/ directory
var KubeflowTemplates = map[Framework]string{
	FrameworkPyTorch:    "pytorchjob.yaml.tmpl", // PyTorchJob CRD
	FrameworkTensorFlow: "tfjob.yaml.tmpl",      // TFJob CRD
	FrameworkXGBoost:    "xgboostjob.yaml.tmpl", // XGBoostJob CRD
	FrameworkMPI:        "mpijob.yaml.tmpl",     // MPIJob CRD (Horovod/DeepSpeed)
}

// Service handles training job operations
type Service struct {
	repo     Repository
	queue    Queue
	executor Executor
}

// NewService creates a new jobs service
func NewService() *Service {
	// TODO: Initialize with:
	// - PostgreSQL repository
	// - SQS queue
	// - K8s executor that renders Kubeflow templates
	//
	// Execution flow:
	// 1. Job submitted → saved to DB with status=pending
	// 2. Job pushed to queue
	// 3. Orchestrator picks up job
	// 4. Render Kubeflow template (KubeflowTemplates[job.Framework])
	// 5. Apply to EKS cluster via kubectl
	// 6. Kubeflow Training Operator manages pods
	// 7. Poll status, update DB
	return &Service{
		repo:     nil,
		queue:    nil,
		executor: nil,
	}
}
