module cloudplane/credential-broker

go 1.21

require (
	github.com/aws/aws-sdk-go-v2 v1.24.0
	github.com/aws/aws-sdk-go-v2/config v1.26.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.5
	github.com/coreos/go-oidc/v3 v3.9.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
)

// Note: Run `go mod tidy` after generating proto files
