# Architecture

cloudplane microservices architecture (MVP: AWS-only).

---

## Overview

```
User Request → Control Plane API → Job Queue → Orchestrator
                                                    │
                                        Credential Broker
                                                    │
                                         (OIDC → AWS STS)
                                                    │
                                                    ↓
                                          User's AWS Account
```

---

## Services

### Credential Broker

Exchanges OIDC tokens for short-lived AWS credentials.

```
services/credential-broker/
├── cmd/server/main.go
├── internal/
│   ├── api/          # HTTP handlers
│   ├── aws/          # STS client
│   ├── oidc/         # Token validation
│   └── authz/        # Authorization
```

**MVP scope**: AWS STS only. GCP/Azure adapters in future.

---

### Control Plane API

User-facing REST API for projects, connections, and jobs.

```
services/control-plane-api/
├── cmd/api/main.go
├── internal/
│   ├── projects/     # Project CRUD
│   ├── connections/  # Cloud account linking
│   └── training/     # Training job submission
```

**MVP scope**: In-memory storage. Database in future.

---

### Orchestrator

Provisions infrastructure and runs Kubernetes workloads.

```
services/orchestrator/
├── cmd/worker/main.go
├── internal/
│   ├── queue/        # Job queue
│   ├── executor/     # Job execution
│   ├── terraform/    # Terraform CLI wrapper
│   └── kubernetes/   # Kubeflow job creation
├── templates/
│   ├── eks-cluster/  # EKS + FSx + networking
│   └── training-jobs/ # Kubeflow job templates
```

**MVP scope**: In-memory queue. SQS/Pub-Sub in future.

---

## Data Flow

```
1. User submits training job request
2. API validates and queues job
3. Orchestrator picks up job
4. Orchestrator calls Credential Broker for AWS credentials
5. Orchestrator runs Terraform (EKS, FSx, networking)
6. Orchestrator creates Kubeflow training job
7. Training runs in user's account
8. User polls for status/logs
```

---

## MVP Limitations

| Component | MVP | Future |
|-----------|-----|--------|
| Cloud | AWS only | + GCP, Azure |
| Storage | In-memory | PostgreSQL |
| Queue | In-memory | SQS |
| Routing | Manual selection | Intelligent |
