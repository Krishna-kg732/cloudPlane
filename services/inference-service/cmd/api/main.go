package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	// TODO: Uncomment after generating proto files
	// pb "cloudplane/inference-service/proto/inferenceservice/v1"
	// "cloudplane/inference-service/internal/server"
)

// Config holds application configuration
type Config struct {
	Port string
}

func loadConfig() *Config {
	return &Config{
		Port: getEnv("GRPC_PORT", "50053"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	cfg := loadConfig()

	// Create gRPC server
	grpcServer := grpc.NewServer(
	// TODO: Add interceptors for:
	// - Logging
	// - Authentication
	// - Metrics
	)

	// TODO: Register service after implementing server
	// svc := server.NewInferenceServer()
	// pb.RegisterInferenceServiceServer(grpcServer, svc)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Start listener
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting inference-service gRPC server on port %s", cfg.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC server...")

	grpcServer.GracefulStop()
	log.Println("Server exited")
}
