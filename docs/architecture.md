# Architecture

This document describes the system architecture for cloudplane.

## Overview

cloudplane uses a microservices architecture with clear service boundaries. **All internal services communicate via gRPC** for performance and type safety, while the **user-facing Control Plane API uses REST**. All services maintain strict separation between the control plane (cloudplane-owned) and execution plane (user-owned).

---

## Service Communication Summary

| From | To | Protocol | Port |
|------|-----|----------|------|
| User/Client | Control Plane API | REST | 8081 |
| Control Plane API | Training Service | **gRPC** | 50052 |
| Control Plane API | Inference Service | **gRPC** | 50053 |
| Orchestrator | Credential Broker | **gRPC** | 50051 |
| Orchestrator | User's Cloud | AWS SDK | — |

```
┌─────────────────────────────────────────────────────────────┐
│                  cloudplane Control Plane                   │
│                                                             │
│  ┌───────────────┐                                          │
│  │ Control Plane │──── REST (user-facing) ────▶ Users       │
│  │ API (:8081)   │                                          │
│  └───────┬───────┘                                          │
│          │ gRPC                                             │
│          ▼                                                  │
│  ┌───────────────┐  ┌───────────────┐  ┌──────────────┐     │
│  │   Training    │  │   Inference   │  │  Credential  │     │
│  │   Service     │  │   Service     │  │   Broker     │     │
│  │   (:50052)    │  │   (:50053)    │  │   (:50051)   │     │
│  └───────────────┘  └───────────────┘  └──────────────┘     │
│          │                  │                 ▲             │
│          ▼                  ▼                 │             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                    Orchestrator                       │   │
│  │             (Terraform + Kubernetes)                  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                         │ OIDC→STS AssumeRole
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              User-Owned Cloud Account (AWS)                 │
└─────────────────────────────────────────────────────────────┘
```

---

## Credential Broker

**Architecture**: gRPC server handling OIDC token exchange for cloud credentials.

**Port**: 50051

### gRPC Service Definition

```protobuf
service CredentialBrokerService {
  rpc IssueAWSCredentials(IssueAWSCredentialsRequest) returns (IssueAWSCredentialsResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

### Example: Calling from Orchestrator

```go
import (
    "google.golang.org/grpc"
    pb "cloudplane/credential-broker/proto/credentialbroker/v1"
)

// Create client
conn, _ := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := pb.NewCredentialBrokerServiceClient(conn)

// Request credentials
resp, err := client.IssueAWSCredentials(ctx, &pb.IssueAWSCredentialsRequest{
    RoleArn: "arn:aws:iam::123456789012:role/DeployRole",
    Ttl:     900,
})

// Use credentials
os.Setenv("AWS_ACCESS_KEY_ID", resp.AccessKeyId)
os.Setenv("AWS_SECRET_ACCESS_KEY", resp.SecretAccessKey)
os.Setenv("AWS_SESSION_TOKEN", resp.SessionToken)
```

---

## Training Service

**Architecture**: gRPC server for distributed training job management.

**Port**: 50052

### gRPC Service Definition

```protobuf
service TrainingService {
  rpc SubmitJob(SubmitJobRequest) returns (SubmitJobResponse);
  rpc GetJob(GetJobRequest) returns (GetJobResponse);
  rpc ListJobs(ListJobsRequest) returns (ListJobsResponse);
  rpc CancelJob(CancelJobRequest) returns (CancelJobResponse);
}
```

### Example: Submitting a Job

```go
client := pb.NewTrainingServiceClient(conn)

resp, err := client.SubmitJob(ctx, &pb.SubmitJobRequest{
    ProjectId:     "proj-123",
    Framework:     "pytorch",
    Image:         "user/training:v1",
    Workers:       4,
    GpusPerWorker: 8,
    Command:       []string{"python", "train.py"},
    DataPath:      "s3://bucket/data",
})

fmt.Printf("Job ID: %s, Status: %s\n", resp.JobId, resp.Status)
```

---

## Inference Service

**Architecture**: gRPC server for LLM inference deployment management.

**Port**: 50053

### gRPC Service Definition

```protobuf
service InferenceService {
  rpc CreateDeployment(CreateDeploymentRequest) returns (CreateDeploymentResponse);
  rpc GetDeployment(GetDeploymentRequest) returns (GetDeploymentResponse);
  rpc ScaleDeployment(ScaleDeploymentRequest) returns (ScaleDeploymentResponse);
  rpc DeleteDeployment(DeleteDeploymentRequest) returns (DeleteDeploymentResponse);
}
```

### Example: Creating a Deployment

```go
client := pb.NewInferenceServiceClient(conn)

resp, err := client.CreateDeployment(ctx, &pb.CreateDeploymentRequest{
    ProjectId:       "proj-123",
    Name:            "llama-70b",
    Model:           "meta-llama/Llama-2-70b",
    Engine:          "vllm",
    Replicas:        2,
    GpusPerReplica:  8,
})

fmt.Printf("Deployment ID: %s\n", resp.DeploymentId)
```

---

## Control Plane API

**Architecture**: RESTful HTTP API (Gin) with gRPC clients to internal services.

**Port**: 8081

### REST Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/projects` | Create project |
| POST | `/v1/projects/:id/connections` | Link cloud account |
| POST | `/v1/training-jobs` | Submit training job |
| POST | `/v1/inference` | Deploy inference |

### Example: Handler calling gRPC

```go
func (h *Handler) SubmitTrainingJob(c *gin.Context) {
    var req TrainingJobRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Call training service via gRPC
    resp, err := h.trainingClient.SubmitJob(c.Request.Context(), &pb.SubmitJobRequest{
        ProjectId: req.ProjectID,
        Framework: req.Framework,
        Workers:   int32(req.Workers),
        Image:     req.Image,
    })
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, gin.H{"job_id": resp.JobId, "status": resp.Status})
}
```

---

## Orchestrator

**Architecture**: Worker-based execution engine with gRPC client.

### Example: Job Execution Pipeline

```go
func (e *Executor) Execute(ctx context.Context, job *Job) error {
    // 1. Get credentials from broker
    creds, err := e.credClient.IssueAWSCredentials(ctx, job.RoleARN, 900)
    if err != nil {
        return err
    }
    setAWSCredentials(creds)

    // 2. Provision EKS if needed
    if !e.terraform.ClusterExists(ctx, job.ClusterName) {
        if err := e.terraform.ApplyCluster(ctx, config); err != nil {
            return err
        }
    }

    // 3. Deploy training job
    manifest := e.renderTemplate(job.Framework, job)
    if err := e.k8s.CreateTrainingJob(ctx, manifest); err != nil {
        return err
    }

    // 4. Watch until complete
    return e.watchJobStatus(ctx, job.ID)
}
```

---

## Why gRPC for Internal Services

| Benefit | Description |
|---------|-------------|
| **Performance** | Binary protocol, ~10x faster than JSON |
| **Type Safety** | Protobuf contracts prevent runtime errors |
| **Code Generation** | Auto-generated clients in any language |
| **Streaming** | Bidirectional streaming for logs |

## Why REST for User-Facing API

| Benefit | Description |
|---------|-------------|
| **Accessibility** | Works from browsers, curl, any HTTP client |
| **Simplicity** | No proto compilation for API consumers |
| **Debugging** | Easy to inspect with standard tools |

---

## Shared Libraries

Located in `libs/`:

- **auth**: JWT/OIDC validation
- **cloud**: AWS SDK wrappers
- **config**: Configuration parsing
- **logging**: Structured logging (slog)
