package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Credentials represents temporary AWS credentials
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	ExpiresAt       time.Time
}

// STSClient wraps AWS STS operations
type STSClient struct {
	client *sts.Client
	region string
}

// NewSTSClient creates a new STS client
func NewSTSClient(ctx context.Context, region string) (*STSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &STSClient{
		client: sts.NewFromConfig(cfg),
		region: region,
	}, nil
}

// AssumeRoleWithWebIdentity exchanges an OIDC token for temporary credentials
func (c *STSClient) AssumeRoleWithWebIdentity(ctx context.Context, roleARN, webIdentityToken string, durationSeconds int32) (*Credentials, error) {
	// TODO: Implement actual STS call
	//
	// Implementation should:
	// 1. Create AssumeRoleWithWebIdentityInput with:
	//    - RoleArn: roleARN
	//    - WebIdentityToken: webIdentityToken
	//    - RoleSessionName: generate unique session name (e.g., "cloudplane-{timestamp}")
	//    - DurationSeconds: durationSeconds (max 900 for web identity)
	// 2. Call c.client.AssumeRoleWithWebIdentity(ctx, input)
	// 3. Extract credentials from result.Credentials
	// 4. Return Credentials struct with AccessKeyId, SecretAccessKey, SessionToken, Expiration
	//
	// Error handling:
	// - Handle AccessDenied (trust policy issue)
	// - Handle MalformedPolicyDocument
	// - Handle ExpiredTokenException (token expired)

	return nil, fmt.Errorf("not implemented")
}
