# Control Plane API

User-facing REST API for cloudplane. Uses gRPC clients to communicate with internal services.

## Overview

Internet-facing service handling user authentication, project management, and routing requests to internal services via gRPC.

## REST API

**Port**: 8081

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/projects` | Create project |
| POST | `/v1/projects/:id/connections` | Link cloud |
| POST | `/v1/training-jobs` | Submit job → gRPC |
| POST | `/v1/inference` | Deploy → gRPC |

## Example Usage

### Handler calling gRPC services

```go
type Handler struct {
    trainingClient   pb.TrainingServiceClient
    inferenceClient  pb.InferenceServiceClient
}

func NewHandler() *Handler {
    trainingConn, _ := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
    inferenceConn, _ := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))

    return &Handler{
        trainingClient:  pb.NewTrainingServiceClient(trainingConn),
        inferenceClient: pb.NewInferenceServiceClient(inferenceConn),
    }
}
```

### Training Jobs Handler

```go
func (h *Handler) SubmitTrainingJob(c *gin.Context) {
    var req TrainingJobRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Call training service via gRPC
    resp, err := h.trainingClient.SubmitJob(c.Request.Context(), &pb.SubmitJobRequest{
        ProjectId: req.ProjectID,
        Framework: req.Framework,
        Workers:   int32(req.Workers),
        Image:     req.Image,
        Command:   req.Command,
    })
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, gin.H{"job_id": resp.JobId, "status": resp.Status})
}
```

### Inference Handler

```go
func (h *Handler) CreateDeployment(c *gin.Context) {
    var req DeploymentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Call inference service via gRPC
    resp, err := h.inferenceClient.CreateDeployment(c.Request.Context(), &pb.CreateDeploymentRequest{
        ProjectId:      req.ProjectID,
        Name:           req.Name,
        Model:          req.Model,
        Engine:         req.Engine,
        Replicas:       int32(req.Replicas),
        GpusPerReplica: int32(req.GPUsPerReplica),
    })
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, gin.H{"deployment_id": resp.DeploymentId, "status": resp.Status})
}
```

### JWT Authentication Middleware

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
            return
        }

        claims, err := validateJWT(strings.TrimPrefix(token, "Bearer "))
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
            return
        }

        c.Set("user_id", claims.Subject)
        c.Set("project_id", claims.ProjectID)
        c.Next()
    }
}
```

## Development

```bash
go run cmd/api/main.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP port | `8081` |
| `TRAINING_SERVICE_ADDR` | Training gRPC addr | `localhost:50052` |
| `INFERENCE_SERVICE_ADDR` | Inference gRPC addr | `localhost:50053` |

## Tech Stack

Go 1.21+, Gin, gRPC (clients), JWT
