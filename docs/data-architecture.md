# Data Architecture

Storage and networking design for high-throughput distributed ML training.

---

## Storage Tiers

| Tier | Technology | Throughput | Use Case |
|------|------------|------------|----------|
| **Hot** | FSx for Lustre | 1+ GB/s per TiB | Active training data |
| **Warm** | S3 | Parallel multipart | Datasets, checkpoints |
| **Config** | EFS | Modest | Configs, small files |

---

## FSx for Lustre

High-performance parallel filesystem linked to S3.

**Configuration**:
```hcl
resource "aws_fsx_lustre_file_system" "training" {
  storage_capacity            = 1200  # GB, minimum for PERSISTENT_2
  deployment_type            = "PERSISTENT_2"
  per_unit_storage_throughput = 250   # MB/s per TiB
  
  # Auto-sync with S3
  import_path = "s3://dataset-bucket/training-data"
  export_path = "s3://dataset-bucket/checkpoints"
  auto_import_policy = "NEW_CHANGED_DELETED"
}
```

**Why Lustre**:
- Parallel I/O across all GPUs
- S3 data repository integration
- Automatic checkpoint sync to S3

---

## S3 Integration

**Dataset Flow**:
```
User uploads to S3 → FSx syncs automatically → Pods read from Lustre mount
```

**Checkpoint Flow**:
```
Pod writes to Lustre → FSx syncs to S3 → Durable checkpoint storage
```

---

## Networking

### EFA (Elastic Fabric Adapter)

RDMA networking for GPU-to-GPU communication.

**Required for**:
- PyTorch DDP with NCCL
- Horovod with MPI
- Any multi-node GPU training

**Instance types with EFA**:
- p4d.24xlarge (8x A100)
- p3dn.24xlarge (8x V100)
- p5.48xlarge (8x H100)

**Configuration**:
```yaml
# EFA annotation in PyTorchJob
annotations:
  vpc.amazonaws.com/efa: "true"
resources:
  limits:
    vpc.amazonaws.com/efa: 4  # Number of EFA devices
    nvidia.com/gpu: 8
```

### Placement Groups

Cluster placement for lowest latency.

```hcl
resource "aws_placement_group" "gpu" {
  name     = "gpu-cluster"
  strategy = "cluster"
}
```

### VPC Endpoints

Avoid NAT gateway bottleneck.

```hcl
# S3 Gateway Endpoint
resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.main.id
  service_name = "com.amazonaws.${var.region}.s3"
}

# ECR Interface Endpoints
resource "aws_vpc_endpoint" "ecr_api" {
  vpc_id            = aws_vpc.main.id
  service_name      = "com.amazonaws.${var.region}.ecr.api"
  vpc_endpoint_type = "Interface"
}
```

---

## Training Pod Storage Mounts

```yaml
volumeMounts:
  - name: training-data
    mountPath: /data
  - name: checkpoints
    mountPath: /checkpoints

volumes:
  - name: training-data
    persistentVolumeClaim:
      claimName: fsx-pvc  # FSx for Lustre CSI
  - name: checkpoints
    persistentVolumeClaim:
      claimName: fsx-pvc
```

---

## Checkpoint Strategy

| Frequency | Location | Purpose |
|-----------|----------|---------|
| Every N steps | FSx (fast) | Fault recovery |
| Every epoch | S3 (durable) | Long-term storage |

**Implementation**:
- Training code writes to `/checkpoints` (Lustre mount)
- FSx auto-exports to S3 on file close
- On failure, new pod reads latest from Lustre

---

## Performance Guidelines

| Dataset Size | FSx Capacity | Expected I/O |
|--------------|--------------|--------------|
| < 500 GB | 1.2 TB | 300 MB/s |
| 500 GB - 2 TB | 2.4 TB | 600 MB/s |
| 2 TB - 10 TB | 4.8+ TB | 1.2+ GB/s |

**Rule of thumb**: Provision FSx capacity ≥ 1.5x dataset size.
