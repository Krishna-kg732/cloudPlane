package serving

import (
	"context"
	"time"
)

// Engine represents the inference engine
type Engine string

const (
	EngineVLLM   Engine = "vllm"
	EngineTGI    Engine = "tgi"
	EngineTriton Engine = "triton"
)

// Deployment represents a model deployment
type Deployment struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	Model        string    `json:"model"`         // HuggingFace model or S3 path
	Engine       Engine    `json:"engine"`        // vllm, tgi, triton
	InstanceType string    `json:"instance_type"` // p4d.24xlarge, g5.xlarge, etc.
	Replicas     int       `json:"replicas"`
	Status       string    `json:"status"` // pending, deploying, running, failed
	Endpoint     string    `json:"endpoint,omitempty"`
	CostPerHour  float64   `json:"cost_per_hour,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Repository interface for deployment storage
type Repository interface {
	Create(d *Deployment) (*Deployment, error)
	Get(id string) (*Deployment, error)
	List(projectID string) ([]*Deployment, error)
	UpdateStatus(id, status, endpoint string) error
	Delete(id string) error
}

// Deployer interface for K8s deployment
type Deployer interface {
	Deploy(ctx context.Context, d *Deployment) error
	Status(ctx context.Context, deploymentID string) (string, error)
	Delete(ctx context.Context, deploymentID string) error
}

// Service handles inference deployments
type Service struct {
	repo     Repository
	deployer Deployer
}

// NewService creates a new serving service
func NewService() *Service {
	// TODO: Initialize with PostgreSQL repo and K8s deployer
	return &Service{
		repo:     nil,
		deployer: nil,
	}
}

// --- K8s Deployer (TODO) ---
//
// type K8sDeployer struct {
//     kubeconfig string
// }
//
// func (d *K8sDeployer) Deploy(ctx context.Context, deployment *Deployment) error {
//     // 1. Select template based on engine (vllm, tgi, triton)
//     // 2. Render template with model, replicas, resources
//     // 3. Apply to cluster: kubectl apply -f
//     // 4. Create Service/Ingress for endpoint
//     return nil
// }
