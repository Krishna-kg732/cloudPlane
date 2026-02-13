package main

import (
	"context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	// TODO: Uncomment after generating proto files
	// pb "cloudplane/credential-broker/proto/credentialbroker/v1"
	// "cloudplane/credential-broker/internal/server"
)

// Config holds application configuration
type Config struct {
	Port         string
	OIDCIssuer   string
	OIDCAudience string
	MaxTTL       int
	AWSRegion    string
}

func loadConfig() *Config {
	return &Config{
		Port:         getEnv("GRPC_PORT", "50051"),
		OIDCIssuer:   getEnv("OIDC_ISSUER", ""),
		OIDCAudience: getEnv("OIDC_AUDIENCE", "credential-broker"),
		MaxTTL:       900, // 15 minutes
		AWSRegion:    getEnv("AWS_REGION", "us-east-1"),
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

	// Validate required config
	if cfg.OIDCIssuer == "" {
		log.Println("WARNING: OIDC_ISSUER not set, JWT validation disabled")
	}

	var verifier *oauth2.TokenSource
	if cfg.OIDCIssuer != "" {
		ctx := context.Background()

		provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuer)
		if err != nil {
			log.Fatalf("Failed to initialize OIDC provider: %v", err)
		}

		verifier = provider.Verifier(&oidc.Config{
			ClientID: cfg.OIDCAudience,
		})
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(
	// TODO: Add interceptors for:
	// - Logging
	// - Authentication (OIDC token validation)
	// - Metrics
	// grpc.UnaryInterceptor(authInterceptor),
	)

	// TODO: Register service after implementing server
	// svc := server.NewCredentialBrokerServer(cfg.OIDCIssuer, cfg.OIDCAudience, cfg.MaxTTL, cfg.AWSRegion)
	// pb.RegisterCredentialBrokerServiceServer(grpcServer, svc)

	// Enable reflection for debugging (disable in production)
	reflection.Register(grpcServer)

	// Start listener
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting gRPC server on port %s", cfg.Port)
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

// TODO: Implement auth interceptor
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// TODO: Implementation should:
	// 1. Skip auth for Health method
	// 2. Extract token from metadata
	// 3. Validate OIDC token
	// 4. Add claims to context
	return handler(ctx, req)
}
