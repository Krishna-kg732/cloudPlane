package credclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	// TODO: Uncomment after generating proto files
	// pb "cloudplane/credential-broker/proto/credentialbroker/v1"
)

// Client is a gRPC client for the credential broker service
type Client struct {
	conn *grpc.ClientConn
	// TODO: Uncomment after generating proto files
	// client pb.CredentialBrokerServiceClient
}

// AWSCredentials represents the returned AWS credentials
type AWSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	ExpiresAt       time.Time
}

// NewClient creates a new credential broker gRPC client
func NewClient(addr string) (*Client, error) {
	// TODO: Use TLS in production
	// creds, err := credentials.NewClientTLSFromFile("ca.pem", "")
	// conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to credential broker: %w", err)
	}

	return &Client{
		conn: conn,
		// TODO: Uncomment after generating proto files
		// client: pb.NewCredentialBrokerServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// IssueAWSCredentials requests temporary AWS credentials from the broker
func (c *Client) IssueAWSCredentials(ctx context.Context, roleARN string, ttl int) (*AWSCredentials, error) {
	// TODO: Implement after generating proto files
	//
	// resp, err := c.client.IssueAWSCredentials(ctx, &pb.IssueAWSCredentialsRequest{
	//     RoleArn: roleARN,
	//     Ttl:     int32(ttl),
	// })
	// if err != nil {
	//     return nil, fmt.Errorf("failed to issue credentials: %w", err)
	// }
	//
	// expiresAt, _ := time.Parse(time.RFC3339, resp.ExpiresAt)
	// return &AWSCredentials{
	//     AccessKeyID:     resp.AccessKeyId,
	//     SecretAccessKey: resp.SecretAccessKey,
	//     SessionToken:    resp.SessionToken,
	//     ExpiresAt:       expiresAt,
	// }, nil

	return nil, fmt.Errorf("not implemented")
}
