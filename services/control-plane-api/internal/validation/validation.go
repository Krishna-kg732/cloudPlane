package validation

import (
	"fmt"
	"regexp"
)

// Validators for input validation

var (
	// AWS ARN pattern
	arnPattern = regexp.MustCompile(`^arn:aws:iam::\d{12}:role/.+$`)

	// S3 path pattern
	s3PathPattern = regexp.MustCompile(`^s3://[a-z0-9][a-z0-9.-]{1,61}[a-z0-9](/.*)?$`)
)

// ValidateRoleARN validates an AWS IAM role ARN
func ValidateRoleARN(arn string) error {
	if arn == "" {
		return fmt.Errorf("role ARN is required")
	}
	if !arnPattern.MatchString(arn) {
		return fmt.Errorf("invalid role ARN format: must be arn:aws:iam::{account}:role/{name}")
	}
	return nil
}

// ValidateS3Path validates an S3 path
func ValidateS3Path(path string) error {
	if path == "" {
		return nil // Optional
	}
	if !s3PathPattern.MatchString(path) {
		return fmt.Errorf("invalid S3 path format: must be s3://bucket/path")
	}
	return nil
}

// ValidateProvider validates cloud provider
func ValidateProvider(provider string) error {
	switch provider {
	case "aws":
		return nil
	case "gcp", "azure":
		return fmt.Errorf("provider %s not yet supported (MVP is AWS-only)", provider)
	default:
		return fmt.Errorf("invalid provider: must be aws, gcp, or azure")
	}
}

// ValidateFramework validates training framework
func ValidateFramework(framework string) error {
	switch framework {
	case "pytorch", "tensorflow", "xgboost", "mpi":
		return nil
	default:
		return fmt.Errorf("invalid framework: must be pytorch, tensorflow, xgboost, or mpi")
	}
}

// ValidateWorkerCount validates worker count
func ValidateWorkerCount(workers int) error {
	if workers < 1 {
		return fmt.Errorf("workers must be at least 1")
	}
	if workers > 64 {
		return fmt.Errorf("workers cannot exceed 64")
	}
	return nil
}

// ValidateGPUCount validates GPU count per worker
func ValidateGPUCount(gpus int) error {
	validCounts := []int{1, 2, 4, 8}
	for _, v := range validCounts {
		if gpus == v {
			return nil
		}
	}
	return fmt.Errorf("gpus_per_worker must be 1, 2, 4, or 8")
}
