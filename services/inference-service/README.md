# Inference Service

A gRPC service for managing LLM inference deployments on Kubernetes.

## Overview

Handles LLM inference deployment lifecycle: creation, scaling, monitoring, and deletion. Supports vLLM, TGI, and Triton inference engines.

## gRPC API

**Port**: 50053

```protobuf
service InferenceService {
  rpc CreateDeployment(CreateDeploymentRequest) returns (CreateDeploymentResponse);
  rpc GetDeployment(GetDeploymentRequest) returns (GetDeploymentResponse);
  rpc ScaleDeployment(ScaleDeploymentRequest) returns (ScaleDeploymentResponse);
  rpc DeleteDeployment(DeleteDeploymentRequest) returns (DeleteDeploymentResponse);
}
```

## Example Usage

### Client (Control Plane API)

```go
conn, _ := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := pb.NewInferenceServiceClient(conn)

// Create deployment
resp, err := client.CreateDeployment(ctx, &pb.CreateDeploymentRequest{
    ProjectId:      "proj-123",
    Name:           "llama-70b",
    Model:          "meta-llama/Llama-2-70b",
    Engine:         "vllm",
    Replicas:       2,
    GpusPerReplica: 8,
})
fmt.Printf("Deployment ID: %s\n", resp.DeploymentId)

// Scale deployment
_, err = client.ScaleDeployment(ctx, &pb.ScaleDeploymentRequest{
    DeploymentId: resp.DeploymentId,
    Replicas:     4,
})
```

### Server Implementation

```go
func (s *Server) CreateDeployment(ctx context.Context, req *pb.CreateDeploymentRequest) (*pb.CreateDeploymentResponse, error) {
    deployment := &Deployment{
        ID:        uuid.New().String(),
        ProjectID: req.ProjectId,
        Name:      req.Name,
        Model:     req.Model,
        Engine:    Engine(req.Engine),
        Replicas:  int(req.Replicas),
        Status:    StatusPending,
    }

    // Save to database
    if err := s.repo.Create(ctx, deployment); err != nil {
        return nil, status.Error(codes.Internal, "failed to create deployment")
    }

    // Enqueue for orchestrator
    if err := s.queue.Enqueue(ctx, deployment); err != nil {
        return nil, status.Error(codes.Internal, "failed to queue deployment")
    }

    return &pb.CreateDeploymentResponse{
        DeploymentId: deployment.ID,
        Status:       string(deployment.Status),
    }, nil
}
```

## Supported Engines

| Engine | Description |
|--------|-------------|
| vLLM | High-throughput LLM serving |
| TGI | HuggingFace Text Generation Inference |
| Triton | NVIDIA multi-model inference server |

## Development

```bash
protoc --go_out=. --go-grpc_out=. proto/inference_service.proto
go run cmd/api/main.go
```

## Tech Stack

Go 1.21+, gRPC, Protocol Buffers
