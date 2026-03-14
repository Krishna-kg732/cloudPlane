# cloudplane
## (currently being scaffolded)

A **multi-cloud control plane** that gives you access to AWS, GCP, and Azure infrastructure from a single interface — while keeping everything in your own accounts.

> **MVP**: Currently AWS-only. GCP and Azure coming in future phases.

## Tech Stack

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)
![Terraform](https://img.shields.io/badge/Terraform-7B42BC?style=for-the-badge&logo=terraform&logoColor=white)
![AWS](https://img.shields.io/badge/AWS-FF9900?style=for-the-badge&logo=amazon-aws&logoColor=white)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         User / Web UI                           │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Control Plane API                          │
│              (Projects, Connections, Routing)                   │
└─────────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────┐     ┌─────────────────┐    ┌─────────────────┐
│  Credential │     │    Training     │    │   Inference     │
│   Broker    │     │    Service      │    │    Service      │
│  (OIDC→STS) │     │ (Kubeflow Jobs) │    │  (vLLM, TGI)    │
└─────────────┘     └─────────────────┘    └─────────────────┘
                               │                    │
                               ▼                    ▼
                    ┌─────────────────────────────────────┐
                    │            Orchestrator             │
                    │    (Terraform + K8s Operations)     │
                    └─────────────────────────────────────┘
                                       │
                           ┌───────────┴───────────┐
                           ▼                       ▼
                   ┌─────────────┐         ┌─────────────┐
                   │  Your AWS   │         │  Your GCP   │ (future)
                   │   Account   │         │   Account   │
                   └─────────────┘         └─────────────┘
```

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
│   ├── credential-broker/     # OIDC → AWS STS exchange
│   ├── control-plane-api/     # Projects, Connections, Routing
│   ├── training-service/      # Distributed training jobs
│   ├── inference-service/     # LLM inference deployments
│   └── orchestrator/          # Terraform + K8s operations
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
