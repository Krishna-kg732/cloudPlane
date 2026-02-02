package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	// TODO: Uncomment after generating proto files
	// pb "cloudplane/training-service/proto/trainingservice/v1"
)

// TrainingServer implements the gRPC TrainingService
type TrainingServer struct {
	// TODO: Uncomment after generating proto files
	// pb.UnimplementedTrainingServiceServer
}

// NewTrainingServer creates a new server instance
func NewTrainingServer() *TrainingServer {
	return &TrainingServer{}
}

// Health returns service health status
func (s *TrainingServer) Health(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Return proper HealthResponse after proto generation
	return nil, nil
}

// SubmitJob creates a new training job
func (s *TrainingServer) SubmitJob(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implementation should:
	//
	// 1. Validate request fields
	// 2. Create job in database
	// 3. Generate Kubeflow job manifest from template
	// 4. Submit to orchestrator queue
	// 5. Return job ID
	//
	// Example:
	// job := &Job{
	//     ID:        uuid.New().String(),
	//     ProjectID: req.GetProjectId(),
	//     Framework: req.GetFramework(),
	//     Image:     req.GetImage(),
	//     Workers:   int(req.GetWorkers()),
	//     Status:    StatusPending,
	// }
	//
	// if err := s.repo.Create(ctx, job); err != nil {
	//     return nil, status.Error(codes.Internal, "failed to create job")
	// }
	//
	// if err := s.queue.Enqueue(ctx, job); err != nil {
	//     return nil, status.Error(codes.Internal, "failed to queue job")
	// }
	//
	// return &pb.SubmitJobResponse{
	//     JobId:  job.ID,
	//     Status: string(job.Status),
	// }, nil

	_ = fmt.Sprintf
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// GetJob returns the status of a training job
func (s *TrainingServer) GetJob(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Fetch job from database and return status
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ListJobs returns all jobs for a project
func (s *TrainingServer) ListJobs(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Query jobs from database with filters
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// CancelJob cancels a running job
func (s *TrainingServer) CancelJob(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Mark job as cancelled and signal orchestrator
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
