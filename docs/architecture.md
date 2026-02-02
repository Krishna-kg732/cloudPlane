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

**Scaffolding**:
```
services/credential-broker/
├── cmd/server/main.go           # gRPC server
├── proto/
│   └── credential_broker.proto  # Service definition
├── internal/
│   ├── server/server.go        # gRPC handlers
│   ├── oidc/oidc.go            # OIDC token validation
│   ├── aws/aws.go              # STS AssumeRoleWithWebIdentity
│   └── authz/authz.go          # Authorization
├── go.mod
└── README.md
```

**gRPC Service**:
```protobuf
service CredentialBrokerService {
  rpc IssueAWSCredentials(IssueAWSCredentialsRequest) returns (IssueAWSCredentialsResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

---

## Training Service

**Architecture**: gRPC server for distributed training job management.

**Port**: 50052

**Scaffolding**:
```
services/training-service/
├── cmd/api/main.go              # gRPC server
├── proto/
│   └── training_service.proto   # Service definition
├── internal/
│   ├── server/server.go         # gRPC handlers
│   └── jobs/jobs.go             # Job models
├── templates/                   # Kubeflow templates
├── go.mod
└── README.md
```

**gRPC Service**:
```protobuf
service TrainingService {
  rpc SubmitJob(SubmitJobRequest) returns (SubmitJobResponse);
  rpc GetJob(GetJobRequest) returns (GetJobResponse);
  rpc ListJobs(ListJobsRequest) returns (ListJobsResponse);
  rpc CancelJob(CancelJobRequest) returns (CancelJobResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

---

## Inference Service

**Architecture**: gRPC server for LLM inference deployment management.

**Port**: 50053

**Scaffolding**:
```
services/inference-service/
├── cmd/api/main.go              # gRPC server
├── proto/
│   └── inference_service.proto  # Service definition
├── internal/
│   ├── server/server.go         # gRPC handlers
│   └── serving/serving.go       # Deployment models
├── templates/                   # vLLM, TGI templates
├── go.mod
└── README.md
```

**gRPC Service**:
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

---

## Control Plane API

**Architecture**: RESTful HTTP API (Gin) with gRPC clients to internal services.

**Port**: 8081

**Scaffolding**:
```
services/control-plane-api/
├── cmd/api/main.go              # HTTP server (Gin)
├── internal/
│   ├── auth/auth.go            # JWT validation
│   ├── projects/projects.go    # Project CRUD
│   ├── connections/connections.go
│   ├── trainingclient/client.go   # gRPC client
│   ├── inferenceclient/client.go  # gRPC client
│   └── validation/validation.go
├── go.mod
└── README.md
```

**REST Endpoints** (user-facing):
- `POST /v1/projects` - Create project
- `POST /v1/projects/:id/connections` - Link cloud account
- `POST /v1/training-jobs` - Submit training job → gRPC to training-service
- `POST /v1/inference` - Deploy inference → gRPC to inference-service

---

## Orchestrator

**Architecture**: Worker-based execution engine with gRPC client.

**Scaffolding**:
```
services/orchestrator/
├── cmd/worker/main.go           # Worker
├── internal/
│   ├── executor/executor.go    # Job execution
│   ├── terraform/terraform.go  # Terraform wrapper
│   ├── kubernetes/kubernetes.go
│   ├── queue/queue.go          # Job queue
│   └── credclient/client.go    # gRPC client to credential broker
├── templates/
├── go.mod
└── README.md
```

---

## Observability

**Architecture**: Metrics/logs collector with read-only access.

**Scaffolding**:
```
services/observability/
├── cmd/collector/main.go
├── internal/
│   ├── metrics/metrics.go
│   ├── logs/logs.go
│   └── costs/costs.go
└── README.md
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
