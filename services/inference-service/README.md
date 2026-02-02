# Inference Service

A gRPC service for managing LLM inference deployments on Kubernetes.

## Overview

The inference-service handles LLM inference deployment lifecycle: creation, scaling, monitoring, and deletion. It supports multiple inference engines including vLLM, TGI, and Triton.

## gRPC API

**Proto file**: `proto/inference_service.proto`

```protobuf
service InferenceService {
  rpc CreateDeployment(CreateDeploymentRequest) returns (CreateDeploymentResponse);
  rpc GetDeployment(GetDeploymentRequest) returns (GetDeploymentResponse);
  rpc ListDeployments(ListDeploymentsRequest) returns (ListDeploymentsResponse);
  rpc DeleteDeployment(DeleteDeploymentRequest) returns (DeleteDeploymentResponse);
  rpc ScaleDeployment(ScaleDeploymentRequest) returns (ScaleDeploymentResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

**Port**: 50053 (default)

## Architecture

```
inference-service/
├── cmd/api/main.go              # gRPC server bootstrap
├── proto/
│   └── inference_service.proto  # Service definition
├── internal/
│   ├── server/server.go         # gRPC handlers
│   └── serving/serving.go       # Deployment models
├── templates/                   # Inference engine templates
│   ├── vllm-deployment.yaml.tmpl
│   └── tgi-deployment.yaml.tmpl
├── go.mod
└── Dockerfile
```

## Supported Engines

| Engine | Description | Use Case |
|--------|-------------|----------|
| vLLM | High-throughput LLM serving | Large-scale inference |
| TGI | HuggingFace Text Generation Inference | HuggingFace models |
| Triton | NVIDIA Triton Inference Server | Multi-model, multi-framework |

## Development

```bash
# Generate proto
protoc --go_out=. --go-grpc_out=. proto/inference_service.proto

# Run
go run cmd/api/main.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `50053` |

## Tech Stack

Go 1.21+, gRPC, Protocol Buffers
