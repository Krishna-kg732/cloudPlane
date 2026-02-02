# Execution Flow

What happens after cloudplane obtains temporary credentials to the user's AWS account.

---

## Overview

```
Credential Broker returns temp credentials
              │
              ▼
┌─────────────────────────────────────────────────────────────┐
│                       ORCHESTRATOR                          │
│                                                             │
│  1. Terraform: Provision infrastructure (EKS, VPC, S3)     │
│  2. Kubernetes: Deploy workloads (training jobs, inference)│
│  3. Monitor: Stream logs, update status                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
              │
              ▼
        User's AWS Account
```

---

## Phase 1: Terraform Provisioning

### Setting AWS Credentials

After receiving credentials from the broker, set them as environment variables:

```go
import "os"

func setAWSCredentials(creds *Credentials) {
    os.Setenv("AWS_ACCESS_KEY_ID", creds.AccessKeyID)
    os.Setenv("AWS_SECRET_ACCESS_KEY", creds.SecretAccessKey)
    os.Setenv("AWS_SESSION_TOKEN", creds.SessionToken)
    os.Setenv("AWS_REGION", "us-east-1")
}
```

### Running Terraform

Using the `terraform-exec` library:

```go
import (
    "github.com/hashicorp/terraform-exec/tfexec"
)

func provisionCluster(workDir string, config ClusterConfig) error {
    // Initialize Terraform
    tf, err := tfexec.NewTerraform(workDir, "/usr/bin/terraform")
    if err != nil {
        return err
    }

    // terraform init
    if err := tf.Init(ctx, tfexec.Upgrade(true)); err != nil {
        return fmt.Errorf("terraform init failed: %w", err)
    }

    // terraform apply
    if err := tf.Apply(ctx); err != nil {
        return fmt.Errorf("terraform apply failed: %w", err)
    }

    // Get outputs
    outputs, err := tf.Output(ctx)
    if err != nil {
        return err
    }
    
    clusterEndpoint := outputs["cluster_endpoint"].Value
    return nil
}
```

### Generated tfvars

```hcl
# Generated from template
cluster_name    = "cloudplane-proj-abc123"
region          = "us-east-1"
gpu_node_count  = 4
gpu_node_type   = "p4d.24xlarge"
fsx_capacity_gb = 1200
efa_enabled     = true
```

---

## Phase 2: Kubernetes Operations

### Getting Kubeconfig

```go
import (
    "github.com/aws/aws-sdk-go-v2/service/eks"
)

func getKubeconfig(ctx context.Context, clusterName, region string) (*rest.Config, error) {
    cfg, _ := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    eksClient := eks.NewFromConfig(cfg)

    // Get cluster info
    result, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
        Name: aws.String(clusterName),
    })
    if err != nil {
        return nil, err
    }

    cluster := result.Cluster
    
    // Build kubeconfig
    return &rest.Config{
        Host: *cluster.Endpoint,
        TLSClientConfig: rest.TLSClientConfig{
            CAData: []byte(*cluster.CertificateAuthority.Data),
        },
        // Use AWS IAM authenticator
        ExecProvider: &clientcmdapi.ExecConfig{
            Command: "aws",
            Args:    []string{"eks", "get-token", "--cluster-name", clusterName},
        },
    }, nil
}
```

### Deploying Training Job

```go
import (
    "k8s.io/client-go/dynamic"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func deployTrainingJob(ctx context.Context, client dynamic.Interface, manifest []byte) error {
    // Parse YAML to unstructured
    obj := &unstructured.Unstructured{}
    if err := yaml.Unmarshal(manifest, obj); err != nil {
        return err
    }

    // Get GVR for PyTorchJob
    gvr := schema.GroupVersionResource{
        Group:    "kubeflow.org",
        Version:  "v1",
        Resource: "pytorchjobs",
    }

    // Create the job
    _, err := client.Resource(gvr).Namespace("default").Create(ctx, obj, metav1.CreateOptions{})
    return err
}
```

### Generated PyTorchJob Manifest

```yaml
apiVersion: kubeflow.org/v1
kind: PyTorchJob
metadata:
  name: training-job-xyz
  namespace: default
spec:
  pytorchReplicaSpecs:
    Master:
      replicas: 1
      template:
        spec:
          containers:
            - name: pytorch
              image: user/training:latest
              command: ["python", "train.py"]
              resources:
                limits:
                  nvidia.com/gpu: 8
    Worker:
      replicas: 3
      template:
        spec:
          containers:
            - name: pytorch
              image: user/training:latest
              command: ["python", "train.py"]
              resources:
                limits:
                  nvidia.com/gpu: 8
```

---

## Phase 3: Monitoring

### Watching Job Status

```go
import (
    "k8s.io/apimachinery/pkg/watch"
)

func watchJobStatus(ctx context.Context, client dynamic.Interface, jobName string) error {
    gvr := schema.GroupVersionResource{
        Group:    "kubeflow.org",
        Version:  "v1",
        Resource: "pytorchjobs",
    }

    watcher, err := client.Resource(gvr).Namespace("default").Watch(ctx, metav1.ListOptions{
        FieldSelector: fmt.Sprintf("metadata.name=%s", jobName),
    })
    if err != nil {
        return err
    }
    defer watcher.Stop()

    for event := range watcher.ResultChan() {
        obj := event.Object.(*unstructured.Unstructured)
        conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
        
        for _, c := range conditions {
            cond := c.(map[string]interface{})
            if cond["type"] == "Succeeded" && cond["status"] == "True" {
                log.Println("Job completed successfully!")
                return nil
            }
            if cond["type"] == "Failed" && cond["status"] == "True" {
                return fmt.Errorf("job failed: %s", cond["reason"])
            }
        }
    }
    return nil
}
```

### Streaming Logs

```go
import (
    "k8s.io/client-go/kubernetes"
    corev1 "k8s.io/api/core/v1"
)

func streamLogs(ctx context.Context, clientset *kubernetes.Clientset, podName, namespace string) error {
    req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
        Follow: true,
    })

    stream, err := req.Stream(ctx)
    if err != nil {
        return err
    }
    defer stream.Close()

    // Stream to stdout or WebSocket
    _, err = io.Copy(os.Stdout, stream)
    return err
}
```

---

## Credential Refresh

For long-running operations (> 15 min):

```go
func executeWithCredRefresh(ctx context.Context, broker *credclient.Client, roleARN string, job func(creds *Credentials) error) error {
    ticker := time.NewTicker(12 * time.Minute)  // Refresh before 15 min expiry
    defer ticker.Stop()

    creds, err := broker.IssueAWSCredentials(ctx, roleARN, 900)
    if err != nil {
        return err
    }
    setAWSCredentials(creds)

    // Run job in background
    errCh := make(chan error, 1)
    go func() {
        errCh <- job(creds)
    }()

    for {
        select {
        case err := <-errCh:
            return err  // Job completed
        case <-ticker.C:
            // Refresh credentials
            newCreds, err := broker.IssueAWSCredentials(ctx, roleARN, 900)
            if err != nil {
                log.Printf("Warning: credential refresh failed: %v", err)
                continue
            }
            setAWSCredentials(newCreds)
            log.Println("Credentials refreshed")
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

---

## Complete Executor Pipeline

```go
func (e *Executor) Execute(ctx context.Context, job *queue.Job) error {
    // 1. Get credentials
    creds, err := e.broker.IssueAWSCredentials(ctx, job.RoleARN, 900)
    if err != nil {
        return fmt.Errorf("credential error: %w", err)
    }
    setAWSCredentials(creds)

    // 2. Check/provision cluster
    clusterName := fmt.Sprintf("cloudplane-%s", job.ProjectID)
    exists, _ := e.terraform.ClusterExists(ctx, clusterName, job.Region)
    if !exists {
        if err := e.terraform.ApplyCluster(ctx, ClusterConfig{
            ClusterName:  clusterName,
            Region:       job.Region,
            GPUNodeCount: job.Workers,
            GPUNodeType:  "p4d.24xlarge",
        }); err != nil {
            return fmt.Errorf("terraform error: %w", err)
        }
    }

    // 3. Get kubeconfig
    k8sClient, err := e.getK8sClient(ctx, clusterName, job.Region)
    if err != nil {
        return fmt.Errorf("kubeconfig error: %w", err)
    }

    // 4. Deploy training job
    manifest := e.renderTemplate(job.Framework, job)
    if err := e.k8s.CreateTrainingJob(ctx, k8sClient, manifest); err != nil {
        return fmt.Errorf("k8s error: %w", err)
    }

    // 5. Watch until complete
    return e.watchJobStatus(ctx, k8sClient, job.ID)
}
```

---

## Summary

| Phase | Library | AWS Service |
|-------|---------|-------------|
| Terraform | `hashicorp/terraform-exec` | EKS, VPC, S3, FSx |
| K8s Config | `aws-sdk-go-v2/eks` | EKS DescribeCluster |
| K8s Deploy | `k8s.io/client-go` | — |
| Monitoring | `k8s.io/client-go` | CloudWatch (optional) |
