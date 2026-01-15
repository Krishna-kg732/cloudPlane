# Architecture

This document describes the system architecture for cloudplane.

## Overview

cloudplane uses a microservices architecture with clear service boundaries, designed as a monorepo with future migration to separate repositories. Services communicate via gRPC and message queues, maintaining strict separation between the control plane (cloudplane-owned) and execution plane (user-owned).

---

## Credential Broker

**Architecture**: gRPC server handling OIDC token exchange for cloud credentials.

**Scaffolding**:
```
services/credential-broker/
├── cmd/server/main.go           # Entry point, gRPC server setup
├── internal/
│   ├── api/api.go              # gRPC service implementation
│   ├── oidc/oidc.go            # OIDC token validation (go-oidc)
│   ├── aws/aws.go              # STS AssumeRoleWithWebIdentity
│   ├── gcp/gcp.go              # Workload Identity (future)
│   ├── azure/azure.go          # Managed Identity (future)
│   └── authz/authz.go          # Authorization checks
├── config/config.yaml
├── Dockerfile
└── README.md
```

**Key Components**:
- **OIDC Validator**: Verifies tokens from cloudplane's identity provider
- **AWS STS Client**: Calls `AssumeRoleWithWebIdentity` with user's role ARN
- **Authorization**: Checks project_id → role_arn mappings
- **Audit Logger**: Records all credential vending operations

**Security Model**:
- Credentials vended on-demand, never persisted
- Internal-only service (not exposed to internet)
- Short-lived credentials (15-60 min TTL)
- All requests logged for audit trail

---

## Orchestrator

**Architecture**: Worker-based execution engine processing jobs from a queue.

**Scaffolding**:
```
services/orchestrator/
├── cmd/worker/main.go           # Worker entry point
├── internal/
│   ├── api/api.go              # gRPC job submission server
│   ├── executor/executor.go    # Job execution engine
│   ├── terraform/terraform.go  # Terraform CLI wrapper
│   ├── kubernetes/kubernetes.go # kubectl/Helm operations
│   ├── queue/queue.go          # Job queue (SQS/RabbitMQ)
│   └── state/state.go          # Terraform state management
├── templates/
│   ├── eks-cluster/            # EKS Terraform templates
│   ├── inference-service/      # Inference deployment templates
│   └── vector-db/              # Vector DB templates
├── Dockerfile
└── README.md
```

**Key Components**:
- **Job Queue Consumer**: Pulls deployment jobs submitted by API
- **Credential Client**: Requests temp credentials from credential broker
- **Terraform Engine**: Generates HCL from templates, executes terraform commands
- **Kubernetes Client**: Applies manifests, Helm charts via client-go
- **State Manager**: Manages Terraform state in user's S3 backend

**Workflow**:
1. Poll queue for deployment jobs
2. Request credentials from broker (gRPC)
3. Generate Terraform/K8s configs from templates
4. Execute operations with streaming logs
5. Update job status in database

---

## Control Plane API

**Architecture**: RESTful HTTP API with JWT authentication.

**Scaffolding**:
```
services/control-plane-api/
├── cmd/api/main.go              # HTTP server entry point
├── internal/
│   ├── auth/auth.go            # JWT validation, API keys
│   ├── projects/projects.go    # Project CRUD operations
│   ├── deployments/deployments.go # Deployment lifecycle
│   ├── connections/connections.go # Cloud connection management
│   └── validation/validation.go # Input validation, quotas
├── Dockerfile
└── README.md
```

**Key Components**:
- **Authentication Middleware**: JWT validation for all requests
- **Project Service**: Manages project → cloud connection mappings
- **Deployment Service**: Accepts deployment requests, submits to queue
- **Connection Service**: Stores project_id → role_arn mappings
- **Validation Layer**: Input sanitization, quota enforcement

**API Endpoints**:
- `POST /v1/projects` - Create project
- `POST /v1/projects/:id/connections` - Link cloud account
- `POST /v1/deployments` - Submit deployment
- `GET /v1/deployments/:id` - Query status/logs
- `GET /v1/projects/:id/resources` - List deployed resources

**Security**:
- Internet-facing, strict input validation
- Never calls cloud APIs directly
- Stores only role ARNs, never credentials
- Rate limiting per user/project

---

## Observability

**Architecture**: Polling-based metrics/logs collector with read-only access.

**Scaffolding**:
```
services/observability/
├── cmd/collector/main.go        # Collector entry point
├── internal/
│   ├── metrics/metrics.go      # CloudWatch/Prometheus collection
│   ├── logs/logs.go            # Log aggregation
│   ├── costs/costs.go          # Cost Explorer integration
│   └── storage/storage.go      # Time-series DB storage
└── README.md
```

**Key Components**:
- **Metrics Collector**: Polls CloudWatch, Prometheus endpoints
- **Log Aggregator**: Streams CloudWatch Logs from user accounts
- **Cost Analyzer**: Queries Cost Explorer API for spend attribution
- **Storage Backend**: Writes to time-series database (InfluxDB/Prometheus)

**Access Pattern**:
- Read-only IAM permissions
- Periodic polling (5-15 min intervals)
- Exposes data via read-only API to control plane
- No alerting or modification capabilities

---

## Service Communication

```
Control Plane API
    ↓ (submit job)
Job Queue (SQS/RabbitMQ)
    ↓ (poll)
Orchestrator
    ↓ (request creds - gRPC)
Credential Broker
    ↓ (STS assume role)
AWS/GCP/Azure
```

**Communication Patterns**:
- API → Queue: Async job submission
- Orchestrator → Credential Broker: Synchronous gRPC
- Orchestrator → User Cloud: Direct API calls with temp credentials
- Observability → User Cloud: Read-only polling

---

## Shared Libraries

Located in `libs/`, used across services:

- **auth**: JWT/OIDC validation utilities
- **cloud**: AWS/GCP/Azure SDK wrappers
- **config**: Configuration parsing (YAML)
- **logging**: Structured logging (slog)

**Design Principle**: Libraries contain only stateless utilities, no business logic or service-specific code.
