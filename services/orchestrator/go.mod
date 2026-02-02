module cloudplane/orchestrator

go 1.21

require (
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
)

// Note: Run `go mod tidy` after implementation to include:
// - k8s.io/client-go (for Kubernetes API)
// - github.com/hashicorp/terraform-exec (for Terraform operations)
// - github.com/aws/aws-sdk-go-v2 (for AWS SDK)
