# Orchestrator

Terraform and kubectl execution service for cloudplane.

## Overview

Worker-based service that executes infrastructure operations in user-owned cloud accounts. Pulls jobs from queue, requests credentials from broker, and executes Terraform/kubectl.

## Architecture

```
Queue → Orchestrator → Credential Broker → AWS → User's Account
```

## Example Usage

### Job Execution Pipeline

```go
func (e *Executor) Execute(ctx context.Context, job *Job) error {
    // 1. Get credentials from broker (gRPC)
    conn, _ := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    client := pb.NewCredentialBrokerServiceClient(conn)

    resp, err := client.IssueAWSCredentials(ctx, &pb.IssueAWSCredentialsRequest{
        RoleArn: job.RoleARN,
        Ttl:     900,
    })
    if err != nil {
        return fmt.Errorf("credential error: %w", err)
    }

    // Set AWS credentials
    os.Setenv("AWS_ACCESS_KEY_ID", resp.AccessKeyId)
    os.Setenv("AWS_SECRET_ACCESS_KEY", resp.SecretAccessKey)
    os.Setenv("AWS_SESSION_TOKEN", resp.SessionToken)

    // 2. Provision EKS if needed
    if !e.terraform.ClusterExists(ctx, job.ClusterName) {
        if err := e.terraform.ApplyCluster(ctx, config); err != nil {
            return err
        }
    }

    // 3. Deploy training job
    manifest := e.renderTemplate(job.Framework, job)
    return e.k8s.Apply(ctx, manifest)
}
```

### Terraform Provisioning

```go
import "github.com/hashicorp/terraform-exec/tfexec"

func (r *Runner) ApplyCluster(ctx context.Context, config ClusterConfig) error {
    tf, err := tfexec.NewTerraform(workDir, "/usr/bin/terraform")
    if err != nil {
        return err
    }

    // terraform init
    if err := tf.Init(ctx, tfexec.Upgrade(true)); err != nil {
        return fmt.Errorf("init failed: %w", err)
    }

    // terraform apply
    if err := tf.Apply(ctx); err != nil {
        return fmt.Errorf("apply failed: %w", err)
    }

    return nil
}
```

### Kubernetes Deployment

```go
import "k8s.io/client-go/dynamic"

func (c *Client) CreateTrainingJob(ctx context.Context, manifest []byte) error {
    obj := &unstructured.Unstructured{}
    yaml.Unmarshal(manifest, obj)

    gvr := schema.GroupVersionResource{
        Group:    "kubeflow.org",
        Version:  "v1",
        Resource: "pytorchjobs",
    }

    _, err := c.dynamicClient.Resource(gvr).Namespace("default").Create(ctx, obj, metav1.CreateOptions{})
    return err
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CREDENTIAL_BROKER_ADDR` | Broker gRPC address | `localhost:50051` |
| `WORKER_ID` | Worker identifier | `worker-1` |
| `POLL_INTERVAL` | Queue poll interval | `5s` |

## Tech Stack

Go 1.21+, gRPC (client), terraform-exec, client-go, AWS SDK v2
