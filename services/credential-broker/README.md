# Credential Broker

A secure gRPC service that issues short-lived cloud credentials. Only accessible by the orchestrator service.

## Overview

Issues temporary AWS credentials using STS AssumeRoleWithWebIdentity with a maximum TTL of 15 minutes.

## gRPC API

**Port**: 50051

```protobuf
service CredentialBrokerService {
  rpc IssueAWSCredentials(IssueAWSCredentialsRequest) returns (IssueAWSCredentialsResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

## Example Usage

### Client (Orchestrator)

```go
conn, _ := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := pb.NewCredentialBrokerServiceClient(conn)

resp, err := client.IssueAWSCredentials(ctx, &pb.IssueAWSCredentialsRequest{
    RoleArn: "arn:aws:iam::123456789012:role/DeployRole",
    Ttl:     900,
})

// Use temporary credentials
os.Setenv("AWS_ACCESS_KEY_ID", resp.AccessKeyId)
os.Setenv("AWS_SECRET_ACCESS_KEY", resp.SecretAccessKey)
os.Setenv("AWS_SESSION_TOKEN", resp.SessionToken)
```

### Server Implementation

```go
func (s *Server) IssueAWSCredentials(ctx context.Context, req *pb.IssueAWSCredentialsRequest) (*pb.IssueAWSCredentialsResponse, error) {
    // 1. Extract JWT from metadata
    md, _ := metadata.FromIncomingContext(ctx)
    token := strings.TrimPrefix(md.Get("authorization")[0], "Bearer ")

    // 2. Validate OIDC token
    claims, err := s.oidcValidator.Validate(ctx, token)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "invalid token")
    }

    // 3. Call AWS STS
    result, err := s.stsClient.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
        RoleArn:          aws.String(req.RoleArn),
        WebIdentityToken: aws.String(token),
        DurationSeconds:  aws.Int32(req.Ttl),
    })
    if err != nil {
        return nil, status.Error(codes.Internal, "STS error")
    }

    return &pb.IssueAWSCredentialsResponse{
        AccessKeyId:     *result.Credentials.AccessKeyId,
        SecretAccessKey: *result.Credentials.SecretAccessKey,
        SessionToken:    *result.Credentials.SessionToken,
        ExpiresAt:       result.Credentials.Expiration.Format(time.RFC3339),
    }, nil
}
```

## Development

```bash
# Generate proto
protoc --go_out=. --go-grpc_out=. proto/credential_broker.proto

# Run
go run cmd/server/main.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `50051` |
| `OIDC_ISSUER` | OIDC provider URL | (required) |
| `MAX_TTL` | Max credential TTL | `900` |

## Tech Stack

Go 1.21+, gRPC, Protocol Buffers, AWS SDK v2, go-oidc
