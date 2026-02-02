package trainingclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	// TODO: Uncomment after generating proto files
	// pb "cloudplane/training-service/proto/trainingservice/v1"
)

// Client is a gRPC client for the training service
type Client struct {
	conn *grpc.ClientConn
	// TODO: Uncomment after generating proto files
	// client pb.TrainingServiceClient
}

// Job represents a training job
type Job struct {
	ID          string
	ProjectID   string
	Framework   string
	Status      string
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

// SubmitJobRequest represents a job submission request
type SubmitJobRequest struct {
	ProjectID     string
	Framework     string
	Image         string
	Workers       int
	GPUsPerWorker int
	Command       []string
	EnvVars       map[string]string
	DataPath      string
}

// NewClient creates a new training service gRPC client
func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to training service: %w", err)
	}

	return &Client{
		conn: conn,
		// TODO: Uncomment after generating proto files
		// client: pb.NewTrainingServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// SubmitJob submits a new training job
func (c *Client) SubmitJob(ctx context.Context, req *SubmitJobRequest) (string, error) {
	// TODO: Implement after generating proto files
	return "", fmt.Errorf("not implemented")
}

// GetJob returns job status
func (c *Client) GetJob(ctx context.Context, jobID string) (*Job, error) {
	// TODO: Implement after generating proto files
	return nil, fmt.Errorf("not implemented")
}

// ListJobs returns jobs for a project
func (c *Client) ListJobs(ctx context.Context, projectID string, limit, offset int) ([]*Job, int, error) {
	// TODO: Implement after generating proto files
	return nil, 0, fmt.Errorf("not implemented")
}

// CancelJob cancels a job
func (c *Client) CancelJob(ctx context.Context, jobID string) error {
	// TODO: Implement after generating proto files
	return fmt.Errorf("not implemented")
}
