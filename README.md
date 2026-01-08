# cloudplane

A Vercel-style control plane for deploying AI/ML workloads inside user-owned cloud accounts using delegated trust (OIDC), Kubernetes, and Terraform.

---

## Overview

**What it is:**
- Deploys AI/ML workloads (inference, training, vector DBs) into **user-owned** AWS/GCP/Azure accounts
- Uses OIDC to obtain short-lived cloud credentials—**never stores access keys**
- Provisions infrastructure via Terraform, orchestrates Kubernetes (AWS EKS first)
- Maintains strict separation: control plane (cloudplane-owned) vs execution plane (user-owned)

**What it solves:**
- No static credentials or secrets shared with cloudplane
- Full user ownership—users can manually intervene anytime
- Cross-cloud standardization with deterministic orchestration
- Enterprise-credible security via delegated trust

**What it is NOT:**
- Not a cloud account provisioning service (BYOC: bring your own cloud)
- Not a data/vector database hosting platform (users own everything)
- Not an autonomous agent (no LLMs making infra decisions)
- Not a replacement for direct cloud access (users retain full `kubectl`/Terraform access)

---

## Core Design Principles

1. **Delegated Trust via OIDC**: Users configure IAM trust policies; cloudplane exchanges OIDC tokens for temporary credentials (15-60 min). Revoke access by deleting the trust policy.

2. **Per-Tenant Cloud Isolation**: Each user's infra lives in their own cloud account—no shared VPCs or clusters.

3. **Control Plane vs Execution Plane**: Control plane (cloudplane services) never runs user workloads; execution plane (user clouds) never stores cloudplane secrets.

4. **Deterministic Orchestration**: All infrastructure changes are declarative, explicit, and auditable. No probabilistic decisions.

5. **User Retains Full Access**: Users can always `kubectl`, `terraform`, or use cloud consoles. Manual intervention is explicitly supported.

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  cloudplane Control Plane                    │
│  ┌───────────────┐  ┌──────────────┐  ┌────────────────┐   │
│  │ Control Plane │─▶│ Credential   │─▶│ Orchestrator   │   │
│  │ API           │  │ Broker       │  │                │   │
│  │               │  │              │  │ - Terraform    │   │
│  │ - Auth        │  │ - OIDC→Cloud │  │ - kubectl      │   │
│  │ - Projects    │  │ - STS assume │  │ - K8s ops      │   │
│  │ - Deployments │  │ - Cred issue │  │                │   │
│  └───────────────┘  └──────────────┘  └────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Observability (read-only metrics, logs, costs)       │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                         │ OIDC→STS AssumeRole
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              User-Owned Cloud Account (AWS)                  │
│  ┌────────────────────────────────────────────────────┐     │
│  │ EKS Cluster (Inference, Training, Vector DBs)      │     │
│  └────────────────────────────────────────────────────┘     │
│  S3 │ RDS │ ECR │ VPCs │ IAM                                │
└─────────────────────────────────────────────────────────────┘
```

**Why services are isolated:**
- **Security boundaries**: Credential broker is the only service touching STS
- **Blast radius**: Orchestrator crashes don't affect API availability
- **Independent scaling**: Different resource needs (API: low latency, Orchestrator: high CPU)
- **Future-proof**: Designed to split into separate repos post-MVP

---

## Repository Structure

```
cloudplane/
├── README.md
├── docs/                        # Architecture, security model, threat model
├── services/                    # Independent deployable services
│   ├── credential-broker/       # OIDC→cloud credential exchange
│   ├── orchestrator/            # Terraform/kubectl executor
│   ├── control-plane-api/       # User-facing REST API
│   └── observability/           # Metrics, logs, cost attribution (MVP-light)
├── libs/                        # Shared utilities (no business logic)
│   ├── auth/                    # JWT validation, OIDC utils
│   ├── cloud/                   # Cloud SDK wrappers
│   ├── config/                  # Config schemas
│   └── logging/                 # Structured logging
├── infra/
│   ├── terraform/               # Platform-owned infra (control plane)
│   └── iam/                     # Example IAM trust policies for users
├── scripts/                     # Dev/ops scripts
└── .github/workflows/           # CI/CD
```

---

## Services

### 1. Credential Broker

**Purpose**: Exchange OIDC tokens for short-lived cloud credentials.

**Structure**:
```
credential-broker/
├── cmd/server/main.go
├── internal/
│   ├── api/                     # gRPC server
│   ├── oidc/                    # Token validation
│   ├── aws/                     # STS AssumeRoleWithWebIdentity
│   ├── gcp/                     # Workload Identity (future)
│   ├── azure/                   # Managed Identity (future)
│   └── authz/                   # Authorization checks
├── config/
├── Dockerfile
└── README.md
```

**Responsibilities**:
- Validate OIDC tokens from cloudplane's identity provider
- Call `sts:AssumeRoleWithWebIdentity` with user's IAM role ARN
- Return temp credentials to orchestrator (never leave internal network)
- Log all credential issuance for audit

**What it NEVER does**:
- Store credentials (vended on-demand, discarded immediately)
- Make infra decisions (orchestrator's job)
- Interact with K8s/Terraform
- Handle user auth (API's job)

**What's stored**: `project_id → role_arn` mappings, audit logs  
**What's NEVER stored**: Access keys, secrets, session tokens, OIDC tokens

---

### 2. Orchestrator

**Purpose**: Execute Terraform and kubectl operations in user clouds.

**Structure**:
```
orchestrator/
├── cmd/worker/main.go
├── internal/
│   ├── api/                     # gRPC job submission
│   ├── executor/                # Job execution engine
│   ├── terraform/               # Terraform wrapper
│   ├── kubernetes/              # kubectl, Helm ops
│   ├── queue/                   # Job queue (SQS, RabbitMQ)
│   └── state/                   # Terraform state mgmt
├── templates/
│   ├── eks-cluster/
│   ├── inference-service/
│   └── vector-db/
├── Dockerfile
└── README.md
```

**Responsibilities**:
- Pull jobs from queue (submitted by API)
- Request temp credentials from credential broker
- Generate Terraform config from templates + user input
- Execute `terraform apply`, `kubectl apply`, etc.
- Stream logs to API; store state in user's S3

**What it NEVER does**:
- Authenticate end users
- Make product decisions (executes API instructions)
- Own Terraform state (lives in user's S3)

**Why separate from API**: Long-running ops (10+ min), different resource needs, failure isolation.

---

### 3. Control Plane API

**Purpose**: User-facing REST API for projects and deployments.

**Structure**:
```
control-plane-api/
├── cmd/api/main.go
├── internal/
│   ├── auth/                    # JWT, API keys
│   ├── projects/                # Project CRUD
│   ├── deployments/             # Deployment submission/status
│   ├── connections/             # Cloud connection mgmt
│   └── validation/              # Input validation, quotas
├── Dockerfile
└── README.md
```

**Responsibilities**:
- Authenticate users
- Manage projects and cloud connections (`project_id → role_arn`)
- Accept deployment requests, submit jobs to orchestrator queue
- Query deployment status/logs
- Enforce rate limits

**What it NEVER does**:
- Call AWS/GCP/Azure APIs directly
- Execute Terraform/kubectl
- Store cloud credentials (only role ARNs)

**Why it never talks to clouds**: Security (internet-facing), credential isolation, scalability, auditability.

---

### 4. Observability (MVP-Light)

**Purpose**: Aggregate metrics, logs, cost data from user accounts.

**Structure**:
```
observability/
├── cmd/collector/main.go
├── internal/
│   ├── metrics/                 # CloudWatch, Prometheus
│   ├── logs/                    # CloudWatch Logs, GCP Logging
│   ├── costs/                   # Cost Explorer, GCP Billing
│   └── storage/                 # Time-series DB
└── README.md
```

**Responsibilities**: Periodic polling for metrics/logs/costs, expose via read-only API.

**What it NEVER does**: Write to user accounts, alerting (out of scope), long-term log storage.

**Why read-only**: Security (compromise can't modify infra), simplicity, minimal IAM permissions.

---

## Shared Libraries (libs/)

**Purpose**: Reusable utilities—**no business logic, no stateful operations**.

**Acceptable**:
- JWT validation, AWS SDK helpers, logging utils, config parsing, error types

**Forbidden**:
- Database models, HTTP handlers, business rules, stateful singletons, cross-service RPC clients

**Rule of thumb**: If it needs a database/queue/external API, it's a service, not a library.

---

## Infrastructure (infra/)

**Two classes**:

1. **Platform-owned** (`infra/terraform/`): Control plane services in cloudplane's AWS account
2. **User-owned** (generated by orchestrator): EKS, S3, RDS in user's account from templates

**Why cloudplane never owns user workloads**:
- User retains control (can `terraform destroy` anytime)
- No lock-in (infra keeps running if user leaves cloudplane)
- Audit trail in user's CloudTrail
- Data never leaves user's account

---

## Security Model

- **Delegated Trust**: OIDC-based, no access keys. Users configure IAM trust policies.
- **Short-Lived Credentials**: 15-60 min expiry via STS. Never persisted.
- **Least Privilege**: Users define minimal IAM permissions. No `AdministratorAccess` required.
- **Immediate Revocation**: Delete trust policy → access denied instantly.
- **Auditability**: AWS CloudTrail logs all operations; cloudplane logs credential requests.

---

## Development Philosophy

**Monorepo now, microservices later**:
- Early velocity (shared CI/CD, atomic changes)
- Services architected for separation (gRPC, no cross-imports, independent Dockerfiles)
- Future: Move to separate repos without refactoring

**Service boundaries**:
- ✅ Stay within service, use gRPC/queues for cross-service, version APIs
- ❌ No shared state, no bypassing APIs, no direct service-to-service imports

---

## Explicit Non-Goals

- ❌ No data hosting (all data in user storage)
- ❌ No managed vector database (can deploy, don't host)
- ❌ No autonomous agents or LLM-driven infra
- ❌ No black-box infra (users always have access)
- ❌ No cloud account creation (BYOC only)
- ❌ Not a general-purpose PaaS (AI/ML infra only)

---

## Summary

cloudplane orchestrates AI/ML deployments in user-owned clouds via OIDC-based delegated trust. No credentials stored, users retain full control, deterministic Terraform/kubectl execution. Designed as a monorepo with clear service boundaries for future microservices migration. MVP: EKS on AWS.

**See also**: [`docs/architecture.md`](docs/architecture.md) | [`docs/security-model.md`](docs/security-model.md) | [`docs/threat-model.md`](docs/threat-model.md)
