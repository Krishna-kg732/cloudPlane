package server

import (
	"context"
	"fmt"
	"time"

	// TODO: Uncomment after generating proto files
	// pb "cloudplane/credential-broker/proto/credentialbroker/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CredentialBrokerServer implements the gRPC CredentialBrokerService
type CredentialBrokerServer struct {
	// TODO: Uncomment after generating proto files
	// pb.UnimplementedCredentialBrokerServiceServer

	oidcIssuer   string
	oidcAudience string
	maxTTL       int
	awsRegion    string
}

// NewCredentialBrokerServer creates a new server instance
func NewCredentialBrokerServer(oidcIssuer, oidcAudience string, maxTTL int, awsRegion string) *CredentialBrokerServer {
	return &CredentialBrokerServer{
		oidcIssuer:   oidcIssuer,
		oidcAudience: oidcAudience,
		maxTTL:       maxTTL,
		awsRegion:    awsRegion,
	}
}

// Health returns service health status
func (s *CredentialBrokerServer) Health(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Return proper HealthResponse after proto generation
	return nil, nil
}

// IssueAWSCredentials exchanges OIDC token for temporary AWS credentials
func (s *CredentialBrokerServer) IssueAWSCredentials(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implementation should:
	//
	// 1. Extract and validate request fields:
	//    roleARN := req.GetRoleArn()
	//    ttl := req.GetTtl()
	//
	// 2. Validate role ARN format:
	//    if !isValidARN(roleARN) {
	//        return nil, status.Error(codes.InvalidArgument, "invalid role ARN")
	//    }
	//
	// 3. Enforce TTL limits:
	//    if ttl <= 0 || int(ttl) > s.maxTTL {
	//        ttl = int32(s.maxTTL)
	//    }
	//
	// 4. Extract OIDC token from context metadata:
	//    md, ok := metadata.FromIncomingContext(ctx)
	//    token := md.Get("authorization")[0]
	//
	// 5. Validate OIDC token:
	//    claims, err := s.oidcValidator.Validate(ctx, token)
	//
	// 6. Call AWS STS:
	//    creds, err := s.stsClient.AssumeRoleWithWebIdentity(ctx, roleARN, token, ttl)
	//
	// 7. Return response:
	//    return &pb.IssueAWSCredentialsResponse{
	//        AccessKeyId:     creds.AccessKeyId,
	//        SecretAccessKey: creds.SecretAccessKey,
	//        SessionToken:    creds.SessionToken,
	//        ExpiresAt:       creds.Expiration.Format(time.RFC3339),
	//    }, nil

	_ = time.RFC3339
	_ = fmt.Sprintf

	return nil, status.Error(codes.Unimplemented, "not implemented")
}
