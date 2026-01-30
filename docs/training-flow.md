# Training Job User Flow

Complete walkthrough of submitting a PyTorch distributed training job.

---

## 1. One-Time Setup: Create IAM Trust

User runs CloudFormation in their AWS account:

```bash
aws cloudformation create-stack \
  --stack-name cloudplane-trust \
  --template-url https://cloudplane.io/setup/aws-oidc.yaml \
  --parameters ParameterKey=ProjectID,ParameterValue=my-ml-project \
  --capabilities CAPABILITY_IAM
```

**What this creates:**
- OIDC Identity Provider trusting cloudplane
- IAM Role with trust policy
- Required permissions (EKS, EC2, S3, FSx)

---

## 2. Create Project

```bash
curl -X POST https://api.cloudplane.io/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "my-ml-project"}'
```

**Response:**
```json
{
  "id": "proj_abc123",
  "name": "my-ml-project",
  "created_at": "2024-01-30T10:00:00Z"
}
```

---

## 3. Link AWS Account

```bash
curl -X POST https://api.cloudplane.io/v1/projects/proj_abc123/connections \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "provider": "aws",
    "role_arn": "arn:aws:iam::123456789012:role/cloudplane-my-ml-project",
    "region": "us-east-1"
  }'
```

**Response:**
```json
{
  "id": "conn_xyz789",
  "project_id": "proj_abc123",
  "provider": "aws",
  "role_arn": "arn:aws:iam::123456789012:role/cloudplane-my-ml-project",
  "region": "us-east-1"
}
```

---

## 4. Submit Training Job

```bash
curl -X POST https://api.cloudplane.io/v1/training/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "project_id": "proj_abc123",
    "framework": "pytorch",
    "image": "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-training:v1",
    "command": ["python", "train.py"],
    "args": ["--epochs", "100", "--batch-size", "64"],
    "workers": 4,
    "gpus_per_worker": 8,
    "efa_enabled": true,
    "storage": {
      "dataset_s3_path": "s3://my-bucket/datasets/imagenet",
      "checkpoint_s3_path": "s3://my-bucket/checkpoints/run-001",
      "fsx_capacity_gb": 1200
    }
  }'
```

**Response:**
```json
{
  "id": "job_def456",
  "project_id": "proj_abc123",
  "framework": "pytorch",
  "status": "pending",
  "workers": 4,
  "gpus_per_worker": 8,
  "created_at": "2024-01-30T10:05:00Z"
}
```

---

## 5. Internal: What Happens Next

### 5.1 Credential Exchange

```
training-service → credential-broker
POST /v1/credentials/aws
{
  "role_arn": "arn:aws:iam::123456789012:role/cloudplane-my-ml-project",
  "ttl": 900
}
```

**credential-broker calls AWS STS:**
```go
// aws.go
result, err := stsClient.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
    RoleArn:          aws.String(roleARN),
    WebIdentityToken: aws.String(oidcToken),
    RoleSessionName:  aws.String("cloudplane-job_def456"),
    DurationSeconds:  aws.Int32(900),
})
```

**Returns temporary credentials (15 min TTL).**

---

### 5.2 Cluster Provisioning

**orchestrator** runs Terraform in user's account:

```hcl
# templates/eks-cluster/main.tf (rendered)
module "eks" {
  cluster_name    = "cloudplane-proj_abc123"
  cluster_version = "1.28"
  
  eks_managed_node_groups = {
    gpu = {
      instance_types = ["p4d.24xlarge"]
      desired_size   = 4
      
      # EFA networking
      network_interfaces = [{
        device_index       = 0
        network_card_index = 0
        interface_type     = "efa"
      }]
    }
  }
}
```

```bash
# Executed by orchestrator
cd /workdir/proj_abc123
terraform init
terraform apply -auto-approve
```

---

### 5.3 Storage Setup

**FSx for Lustre linked to S3:**

```hcl
# templates/eks-cluster/storage.tf (rendered)
resource "aws_fsx_lustre_file_system" "training" {
  storage_capacity = 1200
  subnet_ids       = [module.vpc.private_subnets[0]]
  
  # Auto-import from S3
  import_path      = "s3://my-bucket/datasets/imagenet"
  auto_import_policy = "NEW_CHANGED"
}
```

---

### 5.4 Kubeflow Job Creation

**training-service** renders PyTorchJob template:

```yaml
# templates/pytorchjob.yaml.tmpl (rendered)
apiVersion: kubeflow.org/v1
kind: PyTorchJob
metadata:
  name: job-def456
  namespace: training
spec:
  pytorchReplicaSpecs:
    Master:
      replicas: 1
      template:
        spec:
          containers:
            - name: pytorch
              image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-training:v1
              command: ["python", "train.py"]
              args: ["--epochs", "100", "--batch-size", "64"]
              resources:
                limits:
                  nvidia.com/gpu: 8
              volumeMounts:
                - name: fsx
                  mountPath: /data
                - name: checkpoints
                  mountPath: /checkpoints
          volumes:
            - name: fsx
              persistentVolumeClaim:
                claimName: fsx-pvc
    Worker:
      replicas: 3
      template:
        spec:
          containers:
            - name: pytorch
              image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-training:v1
              resources:
                limits:
                  nvidia.com/gpu: 8
```

**Applied to cluster:**
```bash
kubectl apply -f pytorchjob-job_def456.yaml
```

---

## 6. Monitor Job

```bash
curl https://api.cloudplane.io/v1/training/jobs/job_def456 \
  -H "Authorization: Bearer $TOKEN"
```

**Response (running):**
```json
{
  "id": "job_def456",
  "status": "running",
  "started_at": "2024-01-30T10:10:00Z",
  "workers": 4,
  "gpus_per_worker": 8,
  "estimated_cost_usd": 45.60
}
```

---

## 7. Get Logs

```bash
curl https://api.cloudplane.io/v1/training/jobs/job_def456/logs \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "logs": [
    {"pod": "job-def456-master-0", "message": "Epoch 1/100, Loss: 2.45"},
    {"pod": "job-def456-worker-0", "message": "Syncing gradients..."},
    {"pod": "job-def456-worker-1", "message": "Syncing gradients..."}
  ]
}
```

---

## 8. Job Completes

**Final status:**
```json
{
  "id": "job_def456",
  "status": "completed",
  "started_at": "2024-01-30T10:10:00Z",
  "completed_at": "2024-01-30T14:30:00Z",
  "duration_hours": 4.33,
  "cost_usd": 187.50,
  "checkpoints": ["s3://my-bucket/checkpoints/run-001/epoch_100.pt"]
}
```

---

## Summary Flow

```
User                    cloudplane                         User's AWS
  │                          │                                  │
  │─── Create Project ──────▶│                                  │
  │─── Link AWS ────────────▶│                                  │
  │─── Submit Job ──────────▶│                                  │
  │                          │─── Get OIDC token ──────────────▶│
  │                          │◀── STS credentials (15 min) ─────│
  │                          │─── terraform apply ─────────────▶│
  │                          │                         (EKS, FSx, VPC)
  │                          │─── kubectl apply ───────────────▶│
  │                          │                       (PyTorchJob)
  │◀── status: running ──────│                                  │
  │                          │◀── Training runs on GPUs ────────│
  │◀── status: completed ────│                                  │
  │                          │                                  │
  │   Checkpoints in S3 ◀────────────────────────────────────────│
```
