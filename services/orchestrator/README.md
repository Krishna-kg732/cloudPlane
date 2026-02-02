# Orchestrator

Terraform and kubectl execution service for cloudplane.

## Overview

The orchestrator is a worker-based service that executes infrastructure operations in user-owned cloud accounts. It pulls jobs from a queue, requests temporary credentials from the credential broker via **gRPC**, and executes Terraform/kubectl commands.

## Architecture

```
services/orchestrator/
├── cmd/worker/main.go           # Worker entry point
├── internal/
│   ├── executor/executor.go    # Job execution engine
│   ├── terraform/terraform.go  # Terraform CLI wrapper
│   ├── kubernetes/kubernetes.go # kubectl/client-go operations
│   ├── queue/queue.go          # Job queue (SQS/in-memory)
│   ├── credclient/client.go    # gRPC client for credential broker
│   └── state/state.go          # Terraform state management
├── templates/
│   ├── eks-cluster/            # EKS Terraform templates
│   └── training-jobs/          # Kubeflow job templates
├── go.mod
└── Dockerfile
```

## Workflow

1. Poll queue for deployment jobs
2. Request credentials from broker (**gRPC call**)
3. Generate Terraform/K8s configs from templates
4. Execute operations with streaming logs
5. Update job status in database

## Development

```bash
go mod download
go run cmd/worker/main.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CREDENTIAL_BROKER_ADDR` | Credential broker gRPC address | `localhost:50051` |
| `WORKER_ID` | Worker identifier | `worker-1` |
| `POLL_INTERVAL` | Job queue poll interval | `5s` |

## Tech Stack

Go 1.21+, gRPC (client), Terraform CLI, client-go, AWS SDK v2
