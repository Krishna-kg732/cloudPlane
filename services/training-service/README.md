# Training Service

A gRPC service for managing distributed training jobs on Kubernetes.

## Overview

The training-service handles distributed training job lifecycle: submission, monitoring, and cancellation. It works with the orchestrator to deploy Kubeflow training operators (PyTorchJob, TFJob, XGBoostJob, MPIJob).

## gRPC API

**Proto file**: `proto/training_service.proto`

```protobuf
service TrainingService {
  rpc SubmitJob(SubmitJobRequest) returns (SubmitJobResponse);
  rpc GetJob(GetJobRequest) returns (GetJobResponse);
  rpc ListJobs(ListJobsRequest) returns (ListJobsResponse);
  rpc CancelJob(CancelJobRequest) returns (CancelJobResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

**Port**: 50052 (default)

## Architecture

```
training-service/
├── cmd/api/main.go              # gRPC server bootstrap
├── proto/
│   └── training_service.proto   # Service definition
├── internal/
│   ├── server/server.go         # gRPC handlers
│   └── jobs/jobs.go             # Job models and repository
├── templates/                   # Kubeflow job templates
│   ├── pytorchjob.yaml.tmpl
│   ├── tfjob.yaml.tmpl
│   ├── xgboostjob.yaml.tmpl
│   └── mpijob.yaml.tmpl
├── go.mod
└── Dockerfile
```

## Supported Frameworks

| Framework | Kubeflow Operator | Use Case |
|-----------|-------------------|----------|
| PyTorch | PyTorchJob | DDP training, DeepSpeed |
| TensorFlow | TFJob | TF distributed strategy |
| XGBoost | XGBoostJob | Distributed gradient boosting |
| MPI | MPIJob | Horovod, custom MPI workloads |

## Development

```bash
# Generate proto
protoc --go_out=. --go-grpc_out=. proto/training_service.proto

# Run
go run cmd/api/main.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `50052` |

## Tech Stack

Go 1.21+, gRPC, Protocol Buffers
