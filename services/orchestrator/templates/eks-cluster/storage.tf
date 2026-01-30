# FSx for Lustre - High-throughput storage for training data

# Security group for FSx
resource "aws_security_group" "fsx" {
  name        = "${var.cluster_name}-fsx"
  description = "Security group for FSx Lustre"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "Lustre from VPC"
    from_port   = 988
    to_port     = 988
    protocol    = "tcp"
    cidr_blocks = [module.vpc.vpc_cidr_block]
  }

  ingress {
    description = "Lustre from VPC"
    from_port   = 1021
    to_port     = 1023
    protocol    = "tcp"
    cidr_blocks = [module.vpc.vpc_cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.cluster_name}-fsx"
  })
}

# FSx for Lustre file system
resource "aws_fsx_lustre_file_system" "training" {
  storage_capacity            = var.fsx_capacity_gb
  subnet_ids                  = [module.vpc.private_subnets[0]]
  security_group_ids          = [aws_security_group.fsx.id]
  deployment_type             = "PERSISTENT_2"
  per_unit_storage_throughput = 250 # MB/s per TiB

  # S3 data repository association (if bucket provided)
  dynamic "s3" {
    for_each = var.dataset_s3_bucket != "" ? [1] : []
    content {
      import_path = "s3://${var.dataset_s3_bucket}"
      export_path = "s3://${var.checkpoint_s3_bucket != "" ? var.checkpoint_s3_bucket : var.dataset_s3_bucket}/checkpoints"
    }
  }

  tags = merge(var.tags, {
    Name = "${var.cluster_name}-fsx"
  })
}

# FSx CSI Driver for Kubernetes
resource "helm_release" "fsx_csi_driver" {
  name       = "aws-fsx-csi-driver"
  repository = "https://kubernetes-sigs.github.io/aws-fsx-csi-driver"
  chart      = "aws-fsx-csi-driver"
  namespace  = "kube-system"
  version    = "1.8.0"

  depends_on = [module.eks]
}

# StorageClass for FSx
resource "kubernetes_storage_class" "fsx" {
  metadata {
    name = "fsx-lustre"
  }

  storage_provisioner = "fsx.csi.aws.com"

  parameters = {
    subnetId         = module.vpc.private_subnets[0]
    securityGroupIds = aws_security_group.fsx.id
    deploymentType   = "PERSISTENT_2"
    perUnitStorageThroughput = "250"
  }

  reclaim_policy         = "Delete"
  volume_binding_mode    = "Immediate"
  allow_volume_expansion = true

  depends_on = [helm_release.fsx_csi_driver]
}

# PersistentVolume for the FSx file system
resource "kubernetes_persistent_volume" "fsx" {
  metadata {
    name = "fsx-pv"
  }

  spec {
    capacity = {
      storage = "${var.fsx_capacity_gb}Gi"
    }

    access_modes                     = ["ReadWriteMany"]
    persistent_volume_reclaim_policy = "Retain"
    storage_class_name               = kubernetes_storage_class.fsx.metadata[0].name

    persistent_volume_source {
      csi {
        driver        = "fsx.csi.aws.com"
        volume_handle = aws_fsx_lustre_file_system.training.id

        volume_attributes = {
          dnsname   = aws_fsx_lustre_file_system.training.dns_name
          mountname = aws_fsx_lustre_file_system.training.mount_name
        }
      }
    }
  }

  depends_on = [kubernetes_storage_class.fsx]
}

# PersistentVolumeClaim for training pods
resource "kubernetes_persistent_volume_claim" "fsx" {
  metadata {
    name      = "fsx-pvc"
    namespace = "default"
  }

  spec {
    access_modes       = ["ReadWriteMany"]
    storage_class_name = kubernetes_storage_class.fsx.metadata[0].name
    volume_name        = kubernetes_persistent_volume.fsx.metadata[0].name

    resources {
      requests = {
        storage = "${var.fsx_capacity_gb}Gi"
      }
    }
  }

  depends_on = [kubernetes_persistent_volume.fsx]
}

# Outputs
output "fsx_file_system_id" {
  description = "FSx Lustre file system ID"
  value       = aws_fsx_lustre_file_system.training.id
}

output "fsx_dns_name" {
  description = "FSx Lustre DNS name"
  value       = aws_fsx_lustre_file_system.training.dns_name
}

output "fsx_mount_name" {
  description = "FSx Lustre mount name"
  value       = aws_fsx_lustre_file_system.training.mount_name
}
