variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "gpu_node_count" {
  description = "Number of GPU nodes"
  type        = number
  default     = 2
}

variable "gpu_node_type" {
  description = "Instance type for GPU nodes (must support EFA for distributed training)"
  type        = string
  default     = "p4d.24xlarge" # 8x A100, EFA enabled

  validation {
    condition     = can(regex("^(p3dn\\.24xlarge|p4d\\.24xlarge|p5\\.48xlarge)$", var.gpu_node_type))
    error_message = "GPU node type must be EFA-enabled: p3dn.24xlarge, p4d.24xlarge, or p5.48xlarge"
  }
}

variable "fsx_capacity_gb" {
  description = "FSx for Lustre capacity in GB (minimum 1200)"
  type        = number
  default     = 1200

  validation {
    condition     = var.fsx_capacity_gb >= 1200
    error_message = "FSx capacity must be at least 1200 GB"
  }
}

variable "efa_enabled" {
  description = "Enable EFA networking for distributed training"
  type        = bool
  default     = true
}

variable "dataset_s3_bucket" {
  description = "S3 bucket for training datasets"
  type        = string
  default     = ""
}

variable "checkpoint_s3_bucket" {
  description = "S3 bucket for checkpoints"
  type        = string
  default     = ""
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    ManagedBy = "cloudplane"
  }
}
