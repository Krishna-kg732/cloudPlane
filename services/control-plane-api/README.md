# Control Plane API

User-facing REST API for cloudplane. Uses gRPC clients to communicate with internal services.

## Overview

The control plane API is the internet-facing service that handles user authentication, project management, and routes requests to internal services via gRPC.

## Architecture

```
services/control-plane-api/
├── cmd/api/main.go              # HTTP server (Gin)
├── internal/
│   ├── auth/auth.go            # JWT validation middleware
│   ├── projects/projects.go    # Project CRUD
│   ├── connections/connections.go # Cloud connections
│   ├── trainingclient/client.go   # gRPC client for training-service
│   ├── inferenceclient/client.go  # gRPC client for inference-service
│   └── validation/validation.go
├── go.mod
└── Dockerfile
```

## Communication

| External (Users) | Internal (Services) |
|------------------|---------------------|
| REST/HTTPS | gRPC |
| JSON | Protobuf |
| Port 8081 | Ports 50052, 50053 |

## REST Endpoints

| Method | Endpoint | Description | Internal Call |
|--------|----------|-------------|---------------|
| `GET` | `/health` | Health check | — |
| `POST` | `/v1/projects` | Create project | — |
| `GET` | `/v1/projects` | List projects | — |
| `POST` | `/v1/projects/:id/connections` | Link cloud | — |
| `POST` | `/v1/training-jobs` | Submit job | → training-service (gRPC) |
| `GET` | `/v1/training-jobs/:id` | Get job | → training-service (gRPC) |
| `POST` | `/v1/inference` | Deploy | → inference-service (gRPC) |
| `GET` | `/v1/inference/:id` | Get deployment | → inference-service (gRPC) |

## Development

```bash
go mod download
go run cmd/api/main.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8081` |
| `TRAINING_SERVICE_ADDR` | Training service gRPC address | `localhost:50052` |
| `INFERENCE_SERVICE_ADDR` | Inference service gRPC address | `localhost:50053` |

## Security

- Internet-facing, strict input validation
- JWT authentication required
- Never calls cloud APIs directly
- Stores only role ARNs, never credentials

## Tech Stack

Go 1.21+, Gin (HTTP), gRPC (clients), JWT
