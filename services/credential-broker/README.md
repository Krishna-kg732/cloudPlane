# Credential Broker

A secure gRPC service that issues short-lived cloud credentials. Only accessible by the orchestrator service.

## Overview

Issues temporary AWS credentials using STS AssumeRoleWithWebIdentity with a maximum TTL of 15 minutes. Returns credentials directly from AWS with no persistent state.

**What it does:**
- Issues temporary AWS credentials via gRPC
- Enforces 15-minute maximum TTL
- Returns credentials from cloud provider

**What it does NOT do:**
- Authenticate end users
- Evaluate policies
- Store credentials

**Access:** Orchestrator service only via service-to-service JWT with `aud=credential-broker` and `scope=issue:credentials`.

## gRPC API

**Proto file**: `proto/credential_broker.proto`

```protobuf
service CredentialBrokerService {
  rpc IssueAWSCredentials(IssueAWSCredentialsRequest) returns (IssueAWSCredentialsResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

**Port**: 50051 (default)

## Architecture

```
credential-broker/
├── cmd/server/main.go              # gRPC server bootstrap
├── proto/
│   └── credential_broker.proto     # Service definition
├── internal/
│   ├── server/server.go            # gRPC handlers
│   ├── service/credentials.go      # Business logic
│   ├── aws/aws.go                  # AWS STS adapter
│   ├── authz/authz.go              # Authorization
│   └── oidc/oidc.go                # JWT validation
├── go.mod
└── Dockerfile
```

## Development

**Prerequisites:** Go 1.21+, protoc, protoc-gen-go, protoc-gen-go-grpc

**Generate Proto**
```bash
protoc --go_out=. --go-grpc_out=. proto/credential_broker.proto
```

**Running Locally**
```bash
go mod download
go run cmd/server/main.go
```

**Environment Variables**
| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `50051` |
| `OIDC_ISSUER` | OIDC provider URL | (required) |
| `OIDC_AUDIENCE` | Expected JWT audience | `credential-broker` |
| `MAX_TTL` | Maximum credential TTL (seconds) | `900` |
| `AWS_REGION` | AWS region | `us-east-1` |

## Security

**Authentication:** All RPCs require valid JWT with `aud=credential-broker` and `scope=issue:credentials`

**Credential Handling:**
- Never logged or persisted
- Returned directly from AWS STS
- Maximum 15-minute TTL enforced

## Tech Stack

Go 1.21+, gRPC, Protocol Buffers, AWS SDK v2, go-oidc
