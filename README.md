# cloudplane

A **multi-cloud control plane** that gives you access to AWS, GCP, and Azure infrastructure from a single interface — while keeping everything in your own accounts.

> **MVP**: Currently AWS-only. GCP and Azure coming in future phases.

## Tech Stack

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![gRPC](https://img.shields.io/badge/gRPC-244c5a?style=for-the-badge&logo=grpc&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)
![Terraform](https://img.shields.io/badge/Terraform-7B42BC?style=for-the-badge&logo=terraform&logoColor=white)
![AWS](https://img.shields.io/badge/AWS-FF9900?style=for-the-badge&logo=amazon-aws&logoColor=white)

---

## Services

| Service | Purpose | Port | Protocol |
|---------|---------|------|----------|
| **Control Plane API** | User-facing REST API | 8081 | REST |
| **Training Service** | Manages distributed training jobs | 50052 | gRPC |
| **Inference Service** | Manages LLM deployments | 50053 | gRPC |
| **Credential Broker** | Issues temporary AWS credentials | 50051 | gRPC |
| **Orchestrator** | Executes Terraform/kubectl | worker | gRPC client |

---

## Service Communication

| From | To | What's Passed | Purpose |
|------|-----|---------------|---------|
| User | Control Plane API | `JWT + JSON` | API request |
| Control Plane API | Training Service | `gRPC: project_id, framework, workers` | Submit job |
| Control Plane API | Inference Service | `gRPC: project_id, model, engine` | Create deployment |
| Orchestrator | Credential Broker | `gRPC: role_arn + JWT` | Get AWS credentials |
| Credential Broker | AWS STS | `AssumeRoleWithWebIdentity` | Exchange token for creds |
| Orchestrator | User's AWS | `Terraform + temp creds` | Deploy infrastructure |

---

## Runtime Flow

```
┌─────────────────┐      ┌───────────────────┐      ┌─────────────────┐
│   cloudplane    │      │   Credential      │      │    AWS STS      │
│   Orchestrator  │      │   Broker          │      │                 │
└────────┬────────┘      └─────────┬─────────┘      └────────┬────────┘
         │                         │                          │
         │ 1. "I need AWS creds    │                          │
         │    for role X" + JWT    │                          │
         │────────────────────────▶│                          │
         │                         │                          │
         │                         │ 2. Validate JWT          │
         │                         │                          │
         │                         │ 3. AssumeRoleWithWebIdentity
         │                         │─────────────────────────▶│
         │                         │                          │
         │                         │         4. AWS validates │
         │                         │            trust policy  │
         │                         │                          │
         │                         │◀─────────────────────────│
         │                         │  5. Temp credentials     │
         │                         │     (15 min expiry)      │
         │◀────────────────────────│                          │
         │  6. Return credentials  │                          │
         │                         │                          │
         │ 7. Deploy to user's AWS │                          │
```

---

## Architecture

```
                         REST (HTTPS)
User ─────────────────────────────────────▶ Control Plane API (:8081)
                                                   │
                                          gRPC (internal)
                              ┌────────────────────┼────────────────────┐
                              ▼                    ▼                    ▼
                     Training Service     Inference Service       Job Queue
                        (:50052)              (:50053)
                              │                    │
                              └──────────┬─────────┘
                                         ▼
                                   Orchestrator
                                         │
                                    gRPC (internal)
                                         ▼
                               Credential Broker (:50051)
                                         │
                                   AWS SDK/HTTPS
                                         ▼
                                     AWS STS
                                         │
                                         ▼
                              User's AWS Account
```

---

## What Each Service Stores

| Service | Stores | Never Stores |
|---------|--------|--------------|
| Control Plane API | Projects, role_arn mappings | Credentials, tokens |
| Training Service | Jobs, status, timestamps | Credentials |
| Inference Service | Deployments, endpoints | Credentials |
| Credential Broker | Audit logs only | Credentials (never) |
| Orchestrator | Job state (in-memory) | Credentials (discarded after use) |

---

## AWS Services (MVP)

| Category | Services |
|----------|----------|
| **Compute** | EKS, EC2 GPU (p4d, g5), EFA |
| **Storage** | S3, FSx for Lustre, EFS |
| **Networking** | VPC, VPC Endpoints, Security Groups |
| **Identity** | IAM, STS (OIDC) |
| **Container** | ECR |
| **Monitoring** | CloudWatch, CloudTrail |

---

## Workloads

| Type | Frameworks |
|------|------------|
| **Training** | PyTorch DDP, TensorFlow, XGBoost, Horovod/DeepSpeed |
| **Inference** | vLLM, TGI (HuggingFace), Triton |

---

## Repository Structure

```
cloudplane/
├── services/
│   ├── credential-broker/     # OIDC → AWS STS (gRPC)
│   ├── control-plane-api/     # User-facing REST API
│   ├── training-service/      # Training jobs (gRPC)
│   ├── inference-service/     # LLM inference (gRPC)
│   └── orchestrator/          # Terraform + K8s
├── docs/
└── infra/
    └── iam/                   # IAM trust templates
```

---

## Quick Start

### 1. Set Up IAM Trust

```bash
aws cloudformation create-stack \
  --stack-name cloudplane-trust \
  --template-url https://cloudplane.io/setup/aws-oidc.yaml \
  --parameters ParameterKey=ProjectID,ParameterValue=your-project \
  --capabilities CAPABILITY_IAM
```

### 2. Link Account

```bash
curl -X POST https://api.cloudplane.io/v1/connections \
  -d '{"provider": "aws", "role_arn": "arn:aws:iam::...", "region": "us-east-1"}'
```

### 3. Submit Training Job

```bash
curl -X POST https://api.cloudplane.io/v1/training/jobs \
  -d '{"framework": "pytorch", "workers": 4, "gpus_per_worker": 8}'
```

### 4. Deploy Inference

```bash
curl -X POST https://api.cloudplane.io/v1/inference/deployments \
  -d '{"model": "meta-llama/Llama-2-70b", "engine": "vllm", "replicas": 2}'
```

---

## Docs

- [Architecture](docs/architecture.md)
- [Data Architecture](docs/data-architecture.md)
- [Security Model](docs/security-model.md)
