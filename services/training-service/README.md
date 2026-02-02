# Training Service

A gRPC service for managing distributed training jobs on Kubernetes.

## Overview

Handles distributed training job lifecycle: submission, monitoring, and cancellation. Works with Kubeflow training operators (PyTorchJob, TFJob, XGBoostJob, MPIJob).

## gRPC API

**Port**: 50052

```protobuf
service TrainingService {
  rpc SubmitJob(SubmitJobRequest) returns (SubmitJobResponse);
  rpc GetJob(GetJobRequest) returns (GetJobResponse);
  rpc ListJobs(ListJobsRequest) returns (ListJobsResponse);
  rpc CancelJob(CancelJobRequest) returns (CancelJobResponse);
}
```

## Example Usage

### Client (Control Plane API)

```go
conn, _ := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := pb.NewTrainingServiceClient(conn)

// Submit job
resp, err := client.SubmitJob(ctx, &pb.SubmitJobRequest{
    ProjectId:     "proj-123",
    Framework:     "pytorch",
    Image:         "user/training:v1",
    Workers:       4,
    GpusPerWorker: 8,
    Command:       []string{"python", "train.py"},
    DataPath:      "s3://bucket/data",
})
fmt.Printf("Job ID: %s\n", resp.JobId)

// Get status
status, _ := client.GetJob(ctx, &pb.GetJobRequest{JobId: resp.JobId})
fmt.Printf("Status: %s\n", status.Status)
```

### Server Implementation

```go
func (s *Server) SubmitJob(ctx context.Context, req *pb.SubmitJobRequest) (*pb.SubmitJobResponse, error) {
    job := &Job{
        ID:        uuid.New().String(),
        ProjectID: req.ProjectId,
        Framework: req.Framework,
        Image:     req.Image,
        Workers:   int(req.Workers),
        Status:    StatusPending,
    }

    // Save to database
    if err := s.repo.Create(ctx, job); err != nil {
        return nil, status.Error(codes.Internal, "failed to create job")
    }

    // Enqueue for orchestrator
    if err := s.queue.Enqueue(ctx, job); err != nil {
        return nil, status.Error(codes.Internal, "failed to queue job")
    }

    return &pb.SubmitJobResponse{
        JobId:  job.ID,
        Status: string(job.Status),
    }, nil
}
```

## Supported Frameworks

| Framework | Kubeflow Operator | Use Case |
|-----------|-------------------|----------|
| PyTorch | PyTorchJob | DDP, DeepSpeed |
| TensorFlow | TFJob | TF distributed |
| XGBoost | XGBoostJob | Gradient boosting |
| MPI | MPIJob | Horovod |

## Development

```bash
protoc --go_out=. --go-grpc_out=. proto/training_service.proto
go run cmd/api/main.go
```

## Tech Stack

Go 1.21+, gRPC, Protocol Buffers
