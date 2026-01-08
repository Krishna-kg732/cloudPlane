# cloudplane

**A control plane for deploying AI/ML workloads inside user-owned cloud accounts.**

cloudplane standardizes deployment workflows across cloud providers while preserving full user ownership, isolation, and direct access to underlying cloud resources. It uses delegated trust (OIDC), Kubernetes, and Terraform to orchestrate infrastructure without ever storing static cloud credentials.

---

## Overview

### What cloudplane Is

cloudplane is a control plane that:

- Deploys AI/ML workloads (inference services, training jobs, vector databases) into **user-owned** AWS, GCP, or Azure accounts
- Uses **delegated trust** via OIDC to obtain short-lived cloud credentials on behalf of users
- Provisions infrastructure using Terraform and orchestrates Kubernetes clusters (starting with AWS EKS)
- Provides a unified API for multi-tenant infrastructure management across cloud providers
- Maintains strict separation between control plane (cloudplane-owned) and execution plane (user-owned cloud resources)

### What Problems It Solves

- **No static credentials**: Users never share access keys or long-lived secrets with cloudplane
- **Full user ownership**: Users retain complete control over their cloud accounts and can intervene manually at any time
- **Cross-cloud standardization**: Unified deployment interface regardless of underlying cloud provider
- **Enterprise-grade security**: Delegated trust model acceptable to security teams at regulated organizations
- **Deterministic orchestration**: Predictable, auditable infrastructure changes—no LLMs making decisions about production resources

### What It Does NOT Do

cloudplane is **not**:

- A cloud account provisioning service (users bring their own AWS/GCP/Azure accounts)
- A data hosting platform (all data lives in user-owned storage)
- A managed vector database service (users own the databases; cloudplane just deploys them)
- An autonomous agent (no black-box infrastructure decisions)
- A replacement for direct cloud access (users can always `kubectl` or `terraform` their own resources)

---

## Core Design Principles

### 1. Delegated Trust via OIDC

cloudplane **never stores cloud credentials**. Instead:

- Users configure an IAM trust relationship between their cloud account and cloudplane's OIDC provider
- At deployment time, cloudplane exchanges a short-lived OIDC token for temporary cloud credentials (e.g., AWS STS `AssumeRoleWithWebIdentity`)
- Credentials expire automatically (typically 15-60 minutes)
- Revocation is immediate: delete the IAM trust policy, and cloudplane can no longer access that account

### 2. Per-Tenant Cloud Isolation

- Each user's infrastructure lives in **their own cloud account**
- No shared VPCs, no shared Kubernetes clusters, no cross-tenant blast radius
- cloudplane control plane is multi-tenant; execution plane is strictly single-tenant

### 3. Control Plane vs Execution Plane Separation

- **Control plane**: cloudplane-owned services (API, credential broker, orchestrator) running in cloudplane's infrastructure
- **Execution plane**: User-owned cloud accounts where workloads actually run (EKS clusters, S3 buckets, RDS instances, etc.)
- Control plane never runs user workloads; execution plane never stores cloudplane secrets

### 4. Deterministic Orchestration

- All infrastructure changes are explicit, declarative, and version-controlled
- Terraform plans are generated deterministically from user input
- No LLMs, no "smart agents," no probabilistic infrastructure decisions
- If a deployment fails, it fails predictably with actionable error messages

### 5. User Retains Full Cloud Access

- Users can `kubectl get pods`, `terraform state pull`, or click around in the AWS console anytime
- cloudplane never locks users into proprietary abstractions
- Manual intervention is always possible and explicitly supported

---

## System Architecture Overview

cloudplane consists of four primary services:

```
┌─────────────────────────────────────────────────────────────┐
│                     cloudplane Control Plane                 │
│                   (cloudplane-owned infra)                   │
│                                                              │
│  ┌─────────────────┐  ┌──────────────────┐  ┌────────────┐ │
│  │ Control Plane   │  │ Credential       │  │ Orchestrator│ │
│  │ API             │─▶│ Broker           │─▶│             │ │
│  │                 │  │                  │  │             │ │
│  │ - Auth          │  │ - OIDC→Cloud     │  │ - Terraform │ │
│  │ - Projects      │  │ - STS assume     │  │ - kubectl   │ │
│  │ - Deployments   │  │ - Cred vending   │  │ - K8s ops   │ │
│  └─────────────────┘  └──────────────────┘  └────────────┘ │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Observability (read-only)                            │   │
│  │ - Metrics aggregation                                │   │
│  │ - Cost attribution                                   │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ OIDC→STS AssumeRole
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                User-Owned Cloud Account (AWS)                │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ EKS Cluster                                          │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │ Inference   │  │ Training    │  │ Vector DB   │  │   │
│  │  │ Service     │  │ Job         │  │ (pgvector)  │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  S3 Buckets │ RDS/Aurora │ ECR Registries │ VPCs │ IAM      │
└─────────────────────────────────────────────────────────────┘
```

### Service Responsibilities

| Service | Purpose | Why It's Isolated |
|---------|---------|-------------------|
| **Control Plane API** | User-facing REST API for project/deployment management | Handles authentication and business logic; must never talk to cloud APIs directly |
| **Credential Broker** | Exchanges OIDC tokens for short-lived cloud credentials | Security isolation; credential vending must be auditable and separate from orchestration |
| **Orchestrator** | Executes Terraform/kubectl operations in user cloud accounts | Heavy compute; long-running operations; must be independently scalable |
| **Observability** | Read-only metrics, logs, cost data from user accounts | Different security posture (read-only); different scaling characteristics |

### Why Services Are Separate

- **Security boundaries**: Credential broker is the only service that interacts with STS; it must be tightly scoped and auditable
- **Blast radius containment**: If orchestrator crashes during a Terraform run, the API remains available
- **Independent scaling**: Orchestrator needs different resource limits (CPU, memory, timeout) than API
- **Clear ownership**: Each service has a single, well-defined responsibility
- **Future-proof for microservices**: Designed to be split into separate repos/deployments post-MVP

---

## Repository Structure

```
cloudplane/
├── README.md                    # This file
├── docs/                        # Architecture and design documentation
│   ├── architecture.md          # Detailed system design, data flows, sequence diagrams
│   ├── security-model.md        # OIDC trust setup, IAM policies, threat mitigations
│   └── threat-model.md          # Attack vectors, risk analysis, security assumptions
│
├── services/                    # Independent services (each deployable separately)
│   ├── credential-broker/       # OIDC→cloud credential exchange service
│   ├── orchestrator/            # Terraform/kubectl execution engine
│   ├── control-plane-api/       # User-facing REST API
│   └── observability/           # Metrics, logs, cost attribution (MVP-light)
│
├── libs/                        # Shared libraries (import-only; no business logic)
│   ├── auth/                    # JWT validation, OIDC client utilities
│   ├── cloud/                   # Cloud provider SDK wrappers (AWS, GCP, Azure)
│   ├── config/                  # Shared configuration schemas
│   └── logging/                 # Structured logging, telemetry helpers
│
├── infra/                       # Infrastructure-as-code for cloudplane itself
│   ├── terraform/               # Platform-owned infra (control plane services)
│   └── iam/                     # Example IAM trust policies for users
│
├── scripts/                     # Development and operations scripts
│
└── .github/
    └── workflows/               # CI/CD pipelines
```

### Directory Explanations

- **`docs/`**: Technical documentation for engineers operating or extending cloudplane. Not user-facing marketing material.
- **`services/`**: Each subdirectory is a deployable service with its own `Dockerfile`, `README.md`, and `main.go`. Services communicate via gRPC (internal) or HTTP (external).
- **`libs/`**: Shared Go modules imported by services. These contain **only** reusable utilities—never business logic, never stateful operations.
- **`infra/terraform/`**: Terraform for cloudplane's own infrastructure (ECS/EKS for control plane, RDS for metadata, etc.). This is **not** the Terraform we run in user accounts.
- **`infra/iam/`**: Example IAM policies and trust relationship configurations that users copy into their AWS accounts.
- **`scripts/`**: Bash/Python scripts for local development, testing, and operational tasks (e.g., `scripts/setup-local-env.sh`, `scripts/test-oidc-flow.sh`).

---

## Services Overview

### Credential Broker

**Purpose**: Securely exchange cloudplane-issued OIDC tokens for short-lived cloud provider credentials.

#### Structure

```
services/credential-broker/
├── cmd/
│   └── server/
│       └── main.go              # Service entrypoint
├── internal/
│   ├── api/                     # gRPC server implementation
│   ├── oidc/                    # OIDC token validation and introspection
│   ├── aws/                     # AWS STS AssumeRoleWithWebIdentity logic
│   ├── gcp/                     # GCP Workload Identity Federation (future)
│   ├── azure/                   # Azure Managed Identity (future)
│   └── authz/                   # Authorization checks (which user can access which role)
├── config/                      # Service configuration (YAML schemas)
├── Dockerfile
└── README.md
```

#### Responsibilities

- Validate incoming OIDC tokens (issued by cloudplane's identity provider)
- Look up the user's configured cloud IAM role ARN from metadata store
- Call `sts:AssumeRoleWithWebIdentity` to obtain temporary AWS credentials
- Return credentials to the orchestrator (credentials never leave cloudplane's internal network)
- Log all credential issuance events for audit trail

#### Explicit Non-Responsibilities

- **Does NOT** store credentials (credentials are vended on-demand and discarded immediately)
- **Does NOT** make decisions about what infrastructure to create (that's the orchestrator's job)
- **Does NOT** interact with Kubernetes or Terraform (only IAM/STS)
- **Does NOT** handle user authentication (that's the control plane API's job)

#### How It Works: OIDC→AWS Credential Flow

1. **User configures trust relationship** (one-time setup):
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [{
       "Effect": "Allow",
       "Principal": {
         "Federated": "arn:aws:iam::USER_ACCOUNT:oidc-provider/auth.cloudplane.dev"
       },
       "Action": "sts:AssumeRoleWithWebIdentity",
       "Condition": {
         "StringEquals": {
           "auth.cloudplane.dev:sub": "project:abc123",
           "auth.cloudplane.dev:aud": "aws"
         }
       }
     }]
   }
   ```
   This policy says: "Trust OIDC tokens from `auth.cloudplane.dev` where `sub=project:abc123`."

2. **Orchestrator requests credentials**:
   - Orchestrator calls credential broker gRPC endpoint: `GetAWSCredentials(project_id="abc123")`
   - Credential broker generates a short-lived OIDC token with `sub=project:abc123`, `aud=aws`

3. **Credential broker calls AWS STS**:
   ```go
   stsClient.AssumeRoleWithWebIdentity(&sts.AssumeRoleWithWebIdentityInput{
       RoleArn:          aws.String("arn:aws:iam::USER_ACCOUNT:role/cloudplane-deployer"),
       WebIdentityToken: aws.String(oidcToken),
       RoleSessionName:  aws.String("cloudplane-orchestrator-session"),
   })
   ```

4. **AWS validates and returns credentials**:
   - AWS verifies the OIDC token signature against cloudplane's public JWKS
   - AWS checks the trust policy conditions
   - AWS returns temporary credentials (access key, secret key, session token) valid for 15-60 minutes

5. **Orchestrator receives credentials**:
   - Credentials are used in-memory for Terraform/kubectl operations
   - Credentials are **never logged, never persisted, never transmitted outside cloudplane's internal network**

#### What Is Stored

- **Metadata only**: `project_id → role_arn` mapping in Postgres
- **Audit logs**: Timestamp, project ID, role ARN, success/failure

#### What Is NEVER Stored

- AWS access keys
- AWS secret keys
- Session tokens
- OIDC tokens (except transiently in memory during validation)

#### Security Guarantees

- **Short-lived**: Credentials expire automatically (typically 1 hour max)
- **Least privilege**: Users configure IAM roles with minimal necessary permissions
- **Immediate revocation**: Delete the IAM trust policy → credential broker can no longer assume the role
- **Auditability**: All credential vending events logged; AWS CloudTrail logs all STS API calls
- **No persistence**: Credentials never written to disk or database

---

### Orchestrator

**Purpose**: Execute infrastructure changes (Terraform) and Kubernetes operations (kubectl, Helm) in user-owned cloud accounts.

#### Structure

```
services/orchestrator/
├── cmd/
│   └── worker/
│       └── main.go              # Worker process (pulls jobs from queue)
├── internal/
│   ├── api/                     # gRPC server for job submission
│   ├── executor/                # Job execution engine (runs Terraform, kubectl)
│   ├── terraform/               # Terraform wrapper (plan, apply, destroy)
│   ├── kubernetes/              # kubectl and Helm operations
│   ├── queue/                   # Job queue integration (RabbitMQ, SQS, etc.)
│   └── state/                   # Terraform state management
├── templates/                   # Terraform module templates for deployments
│   ├── eks-cluster/             # EKS cluster provisioning
│   ├── inference-service/       # Kubernetes deployment + service + ingress
│   └── vector-db/               # RDS/Aurora with pgvector
├── config/
├── Dockerfile
└── README.md
```

#### Responsibilities

- Pull deployment jobs from a queue (submitted by control plane API)
- Request short-lived credentials from credential broker
- Generate Terraform configuration from templates + user input
- Execute `terraform plan`, `terraform apply`, `terraform destroy`
- Execute `kubectl apply`, `kubectl delete`, `helm install`, etc.
- Stream logs and status updates back to control plane API
- Store Terraform state in user-owned cloud storage (S3, GCS, Azure Blob)

#### Explicit Non-Responsibilities

- **Does NOT** authenticate end users (only authenticates internally to credential broker)
- **Does NOT** make product decisions about what to deploy (executes instructions from API)
- **Does NOT** own Terraform state (state lives in user's S3 bucket)
- **Does NOT** run as a singleton (designed to scale horizontally with multiple workers)

#### How It Works: Deployment Flow

1. **Job submission**:
   - User calls control plane API: `POST /projects/abc123/deployments`
   - API validates request, writes job to queue: `{type: "deploy_inference", project_id: "abc123", ...}`

2. **Job execution**:
   - Orchestrator worker pulls job from queue
   - Worker calls credential broker: `GetAWSCredentials(project_id="abc123")`
   - Worker receives temporary AWS credentials

3. **Terraform generation**:
   - Worker loads template from `templates/inference-service/`
   - Worker injects user parameters (model name, instance type, replicas, etc.)
   - Worker writes generated `.tf` files to ephemeral working directory

4. **Terraform execution**:
   ```bash
   export AWS_ACCESS_KEY_ID=<temp_key>
   export AWS_SECRET_ACCESS_KEY=<temp_secret>
   export AWS_SESSION_TOKEN=<temp_token>
   terraform init -backend-config="bucket=user-terraform-state"
   terraform plan -out=plan.tfplan
   terraform apply plan.tfplan
   ```

5. **Result reporting**:
   - Worker streams logs to control plane API
   - Worker updates job status: `running → succeeded | failed`
   - Worker discards credentials and working directory

#### Why Orchestrator Is Separate from API

- **Long-running operations**: Terraform apply can take 10+ minutes; API must respond quickly
- **Resource isolation**: Terraform/kubectl need high CPU/memory; API needs low-latency network
- **Horizontal scaling**: Can run 10 orchestrator workers, 3 API instances independently
- **Failure isolation**: Terraform crash doesn't bring down the API

---

### Control Plane API

**Purpose**: User-facing REST API for managing projects, cloud connections, and deployments.

#### Structure

```
services/control-plane-api/
├── cmd/
│   └── api/
│       └── main.go              # HTTP server entrypoint
├── internal/
│   ├── auth/                    # User authentication (JWT, API keys)
│   ├── projects/                # Project CRUD operations
│   ├── deployments/             # Deployment submission, status queries
│   ├── connections/             # Cloud connection management (register IAM roles)
│   └── validation/              # Input validation, quota checks
├── config/
├── Dockerfile
└── README.md
```

#### Responsibilities

- Authenticate users (JWT tokens, API keys)
- Manage user accounts and projects
- Register cloud connections (store `project_id → role_arn` mappings)
- Accept deployment requests and submit jobs to orchestrator queue
- Query deployment status and logs
- Enforce rate limits and quotas

#### Explicit Non-Responsibilities

- **Does NOT** call AWS/GCP/Azure APIs directly (delegates to orchestrator via credential broker)
- **Does NOT** execute Terraform or kubectl (that's orchestrator's job)
- **Does NOT** store cloud credentials (only stores role ARNs and project metadata)
- **Does NOT** implement business logic for infrastructure creation (uses orchestrator templates)

#### Key Flows

**Registering a Cloud Connection**:
```http
POST /projects/abc123/connections
{
  "provider": "aws",
  "role_arn": "arn:aws:iam::123456789012:role/cloudplane-deployer",
  "region": "us-west-2"
}
```
- API validates IAM role ARN format
- API stores mapping in database: `project_abc123 → role_arn`
- API returns connection ID
- **API never calls AWS** (orchestrator will use this role_arn later)

**Submitting a Deployment**:
```http
POST /projects/abc123/deployments
{
  "type": "inference_service",
  "model": "mistralai/Mistral-7B-Instruct-v0.2",
  "instance_type": "g5.xlarge",
  "replicas": 2
}
```
- API validates input against schema
- API checks project quotas
- API writes job to queue: `{project_id: "abc123", type: "inference_service", ...}`
- API returns deployment ID
- Orchestrator picks up job asynchronously

#### Why This Service Never Talks to Cloud APIs

- **Security**: API is internet-facing; orchestrator is internal-only. If API is compromised, attacker still can't access user clouds.
- **Credential isolation**: Only orchestrator+credential broker handle cloud credentials; API never sees them.
- **Scalability**: API serves 1000s of req/sec; cloud operations are 10s of ops/min. Different scaling profiles.
- **Auditability**: Clear separation between "user said X" (API) and "we did X" (orchestrator).

---

### Observability Service (MVP-Light)

**Purpose**: Aggregate metrics, logs, and cost data from user-owned cloud accounts for display in cloudplane UI.

#### Structure

```
services/observability/
├── cmd/
│   └── collector/
│       └── main.go              # Metric/log collector
├── internal/
│   ├── metrics/                 # Prometheus/CloudWatch metric collection
│   ├── logs/                    # CloudWatch Logs / GCP Logging queries
│   ├── costs/                   # AWS Cost Explorer / GCP Billing API
│   └── storage/                 # Time-series DB (Prometheus, InfluxDB)
├── config/
├── Dockerfile
└── README.md
```

#### Responsibilities

- Periodically query user cloud accounts for metrics (CPU, memory, request latency)
- Aggregate logs from Kubernetes clusters (via CloudWatch, Fluent Bit, etc.)
- Query cloud billing APIs for per-project cost attribution
- Expose metrics via read-only API for cloudplane UI

#### Explicit Non-Responsibilities

- **Does NOT** write to user cloud accounts (strictly read-only)
- **Does NOT** implement alerting or anomaly detection (out of scope for MVP)
- **Does NOT** store raw logs long-term (links to CloudWatch/GCP Logging)
- **Does NOT** replace native cloud monitoring (users should still use CloudWatch/GCP Monitoring)

#### Why This Is Read-Only

- **Security**: If observability service is compromised, attacker cannot modify infrastructure
- **Simplicity**: No complex state management; just periodic polling
- **User trust**: Users can grant minimal IAM permissions (`cloudwatch:GetMetricData`, `logs:FilterLogEvents`) without deployment privileges

#### Why Advanced Observability Is Out of Scope

- **MVP focus**: Core value is deployment, not monitoring
- **Native tools are good**: AWS CloudWatch, GCP Monitoring, Datadog already exist and work well
- **Avoid over-engineering**: Building a competitive observability platform is a separate product
- **Future**: Integrate with existing tools (Datadog, New Relic, Grafana Cloud) rather than reinventing

---

## Shared Libraries (libs/)

### Purpose

Shared libraries contain **reusable utilities** imported by multiple services. They must **never** contain business logic, stateful operations, or service-specific code.

### What Belongs in Shared Libraries

✅ **Acceptable**:
- OIDC token validation logic (JWT signature verification)
- AWS SDK client initialization helpers
- Structured logging utilities (log level, JSON formatting)
- Configuration schema parsing (YAML → Go structs)
- Common error types and error handling utilities
- HTTP middleware (rate limiting, request ID injection)

### What Must NEVER Go in Shared Libraries

❌ **Forbidden**:
- Database models or ORM logic (each service owns its schema)
- HTTP route handlers (service-specific logic)
- Business rules (e.g., "users can have max 5 projects")
- Stateful singletons (e.g., global database connection pools)
- Cross-service RPC clients (services should not import each other)

### Examples of Acceptable Shared Code

**`libs/auth/oidc.go`**:
```go
// ValidateJWT verifies a JWT token signature and extracts claims.
// This is pure, stateless logic suitable for a shared library.
func ValidateJWT(token string, jwksURL string) (*Claims, error) {
    // ...
}
```

**`libs/cloud/aws.go`**:
```go
// NewSTSClient creates an AWS STS client with standard retry/timeout config.
// This is a utility wrapper, not business logic.
func NewSTSClient(region string) *sts.STS {
    // ...
}
```

### Why Business Logic Must Not Live Here

- **Shared libraries create coupling**: If service A and service B both import `libs/deployments`, they're now tightly coupled
- **Business logic changes frequently**: Shared code is hard to version and migrate
- **Testing becomes difficult**: Mocking shared stateful code is error-prone
- **Services should be independently deployable**: Shared business logic violates this principle

**Rule of thumb**: If code needs access to a database, message queue, or external API, it belongs in a service, not a library.

---

## Infrastructure Code (infra/)

### Platform-Owned Infrastructure vs User-Owned Infrastructure

cloudplane manages **two distinct classes of infrastructure**:

1. **Platform-owned infrastructure** (`infra/terraform/`):
   - The control plane itself (API servers, databases, message queues)
   - Runs in cloudplane's AWS account
   - Managed by cloudplane operators via CI/CD

2. **User-owned infrastructure** (generated dynamically by orchestrator):
   - EKS clusters, S3 buckets, RDS instances in user's AWS account
   - Generated from templates in `services/orchestrator/templates/`
   - Executed by orchestrator with user's temporary credentials

### Terraform Usage Philosophy

**For cloudplane's own infrastructure**:
- Standard Terraform workflow: `terraform plan` → review → `terraform apply`
- State stored in cloudplane's own S3 bucket with locking
- Changes deployed via GitHub Actions or manual operations

**For user infrastructure**:
- Terraform modules are **templates**, not directly applied
- Orchestrator generates `.tf` files dynamically from user input
- State stored in **user's S3 bucket** (cloudplane never owns user state)
- Orchestrator applies changes with temporary credentials

### Why cloudplane Never Owns User Workloads

- **User retains control**: User can `terraform destroy` their own infrastructure anytime
- **No lock-in**: If user stops using cloudplane, their infrastructure keeps running
- **Auditability**: All changes logged in user's CloudTrail, not cloudplane's
- **Compliance**: User's data never leaves their account; cloudplane is just an orchestration tool

### Example: infra/iam/

Contains example IAM policies that users copy into their AWS accounts:

**`infra/iam/aws-trust-policy.json`**:
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Federated": "arn:aws:iam::USER_ACCOUNT:oidc-provider/auth.cloudplane.dev"
    },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "auth.cloudplane.dev:sub": "project:PROJECT_ID"
      }
    }
  }]
}
```

**`infra/iam/aws-deployer-policy.json`**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "eks:*",
        "ec2:*",
        "s3:*",
        "iam:PassRole"
      ],
      "Resource": "*"
    }
  ]
}
```

Users customize these policies based on their security requirements (e.g., restrict to specific regions, VPCs, or resource tags).

---

## Security Model Summary

### Delegated Trust

- cloudplane **never asks users for AWS access keys**
- Users configure IAM trust policies in their AWS accounts
- cloudplane proves its identity via OIDC, AWS validates and grants temporary access
- Users can revoke access by deleting the trust policy—no coordination with cloudplane required

### Short-Lived Credentials

- All cloud credentials obtained via STS `AssumeRoleWithWebIdentity`
- Credentials expire after 15-60 minutes (configurable by user's IAM policy)
- No credential refresh/renewal—orchestrator re-authenticates for each job
- Credentials never written to disk, never logged, never transmitted outside cloudplane's internal network

### Least Privilege

- Users define IAM policies with minimum necessary permissions
- cloudplane provides example policies but does not require blanket `AdministratorAccess`
- Recommended: Separate roles for different deployment types (e.g., `cloudplane-eks-deployer`, `cloudplane-s3-manager`)
- Users can scope permissions to specific resource tags, VPCs, or regions

### Revocation Behavior

- **Immediate**: Delete IAM trust policy → next credential request fails
- **Automatic expiry**: Even if trust policy still exists, credentials expire within 1 hour
- **No persistent sessions**: cloudplane never maintains long-lived sessions to user accounts

### Auditability via Native Cloud Logs

- **AWS CloudTrail**: Logs all STS `AssumeRoleWithWebIdentity` calls and subsequent API operations
- **cloudplane audit logs**: Logs all credential requests (timestamp, project ID, role ARN, success/failure)
- **No hidden operations**: Every action cloudplane takes in user's account is visible in CloudTrail
- **User owns the logs**: CloudTrail logs live in user's S3 bucket, not cloudplane's

---

## Development Philosophy

### Why This Starts as a Monorepo

- **Early-stage velocity**: Monorepo reduces boilerplate (shared CI/CD, dependency management)
- **Atomic changes**: Can update shared libraries and services in a single commit
- **Easier local development**: Clone one repo, run `docker-compose up`, all services available
- **Single source of truth**: Documentation, architecture decisions, and code in one place

### How Services Are Designed for Future Separation

Even though this is a monorepo, services are architected as if they were separate repos:

- **No direct imports between services**: Service A cannot `import "cloudplane/services/service-b"`
- **gRPC for inter-service communication**: Services talk over network, not function calls
- **Independent Dockerfiles**: Each service has its own container image
- **Separate configuration**: Each service reads its own config file (no global config singleton)
- **Clear API boundaries**: Each service exposes a versioned API (gRPC or HTTP)

**Future migration path**: When team/codebase grows, move `services/credential-broker/` to a separate repo without architectural changes.

### How Contributors Should Respect Service Boundaries

- **Do NOT** add shared state between services (e.g., shared in-memory cache)
- **Do NOT** bypass APIs (e.g., orchestrator directly querying API's database)
- **Do** use message queues or gRPC for cross-service communication
- **Do** version APIs and maintain backward compatibility
- **Do** write service-specific tests that don't depend on other services running

**Code review checklist**:
- ✅ Does this change stay within a single service boundary?
- ✅ If it touches multiple services, does it use a versioned API?
- ✅ Can this service be deployed independently?
- ❌ Does this add a direct dependency between services?

---

## Explicit Non-Goals

cloudplane is **intentionally scoped** to avoid mission creep. The following are **explicit non-goals**:

### No Data Hosting

- cloudplane does not store user data (training datasets, model weights, inference results)
- All data lives in user-owned S3/GCS/Azure Blob Storage
- cloudplane only stores metadata (project names, deployment configs, role ARNs)

### No Vector Database Hosting

- cloudplane can **deploy** vector databases (pgvector, Pinecone, Weaviate) in user accounts
- cloudplane does **not** run a managed vector database service
- Users own the database; cloudplane just automates provisioning

### No Autonomous Agents

- cloudplane does not use LLMs to make infrastructure decisions
- All infrastructure changes are explicit, declarative, and user-initiated
- No "AI assistant that provisions EC2 instances based on vibes"

### No Black-Box Infrastructure

- Users can always inspect Terraform state, kubectl their clusters, SSH into nodes
- cloudplane provides abstraction, not obfuscation
- If cloudplane generates bad Terraform, users can fix it manually

### No Cloud Account Creation for Users

- cloudplane assumes users already have AWS/GCP/Azure accounts
- cloudplane does not integrate with AWS Organizations, GCP Billing, or Azure Enrollment APIs
- Users bring their own accounts; cloudplane connects to them

### No "Vercel for Everything"

- cloudplane is scoped to AI/ML infrastructure (model inference, training, vector DBs)
- cloudplane is **not** a general-purpose PaaS for web apps, databases, or static sites
- If you want to deploy a Next.js app, use Vercel

---

## Summary

cloudplane is a control plane for deploying AI/ML workloads in user-owned cloud accounts. It uses delegated trust (OIDC) to obtain short-lived credentials, executes Terraform and kubectl operations on behalf of users, and maintains strict separation between control plane (cloudplane-owned) and execution plane (user-owned). The system is designed with security, auditability, and user ownership as first-class principles.

**Key Takeaways**:
- No static credentials ever stored
- Users retain full control over their cloud accounts
- Deterministic, predictable infrastructure orchestration
- Designed as a monorepo but architected for future service separation
- MVP-first: focused on EKS, AWS, and core deployment workflows

For architecture details, see [`docs/architecture.md`](docs/architecture.md).  
For security model, see [`docs/security-model.md`](docs/security-model.md).  
For threat analysis, see [`docs/threat-model.md`](docs/threat-model.md).
