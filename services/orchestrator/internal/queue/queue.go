package queue

import (
	"fmt"
	"time"
)

// JobStatus represents job status
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
)

// Job represents a job in the queue
type Job struct {
	ID            string
	ProjectID     string
	Framework     string // pytorch, tensorflow, xgboost, mpi
	Image         string
	Command       []string
	Args          []string
	Workers       int
	GPUsPerWorker int
	EFAEnabled    bool
	RoleARN       string
	Region        string
	Storage       StorageConfig
	Status        JobStatus
	Error         string
	CreatedAt     time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
}

// StorageConfig for the job
type StorageConfig struct {
	DatasetS3Path    string
	CheckpointS3Path string
	FSxCapacityGB    int
}

// Queue interface for job queues
type Queue interface {
	Push(job *Job) error
	Poll() (*Job, error)
	MarkRunning(id string) error
	MarkCompleted(id string) error
	MarkFailed(id string, errMsg string) error
}

// InMemoryQueue is a simple in-memory queue for MVP/testing
type InMemoryQueue struct {
	jobs []*Job
}

// NewInMemoryQueue creates a new in-memory queue
func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		jobs: make([]*Job, 0),
	}
}

// Push adds a job to the queue
func (q *InMemoryQueue) Push(job *Job) error {
	job.Status = StatusPending
	job.CreatedAt = time.Now()
	q.jobs = append(q.jobs, job)
	return nil
}

// Poll returns the next pending job
func (q *InMemoryQueue) Poll() (*Job, error) {
	for _, job := range q.jobs {
		if job.Status == StatusPending {
			job.Status = StatusRunning
			now := time.Now()
			job.StartedAt = &now
			return job, nil
		}
	}
	return nil, nil // No job available, not an error
}

// MarkRunning marks a job as running
func (q *InMemoryQueue) MarkRunning(id string) error {
	for _, job := range q.jobs {
		if job.ID == id {
			job.Status = StatusRunning
			now := time.Now()
			job.StartedAt = &now
			return nil
		}
	}
	return fmt.Errorf("job not found: %s", id)
}

// MarkCompleted marks a job as completed
func (q *InMemoryQueue) MarkCompleted(id string) error {
	for _, job := range q.jobs {
		if job.ID == id {
			job.Status = StatusCompleted
			now := time.Now()
			job.CompletedAt = &now
			return nil
		}
	}
	return fmt.Errorf("job not found: %s", id)
}

// MarkFailed marks a job as failed
func (q *InMemoryQueue) MarkFailed(id string, errMsg string) error {
	for _, job := range q.jobs {
		if job.ID == id {
			job.Status = StatusFailed
			job.Error = errMsg
			now := time.Now()
			job.CompletedAt = &now
			return nil
		}
	}
	return fmt.Errorf("job not found: %s", id)
}

// --- SQS Queue (TODO for production) ---
//
// type SQSQueue struct {
//     client *sqs.Client
//     url    string
// }
//
// func NewSQSQueue(queueURL string) (*SQSQueue, error) {
//     cfg, err := config.LoadDefaultConfig(context.Background())
//     if err != nil {
//         return nil, err
//     }
//     return &SQSQueue{
//         client: sqs.NewFromConfig(cfg),
//         url:    queueURL,
//     }, nil
// }
