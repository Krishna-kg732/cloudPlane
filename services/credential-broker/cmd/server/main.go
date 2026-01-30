package main

<<<<<<< Updated upstream
func main() {
	// Credential broker server entry point
=======
import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudplane/credential-broker/internal/api"
	"cloudplane/credential-broker/internal/authz"
	"cloudplane/credential-broker/internal/aws"
	"cloudplane/credential-broker/internal/oidc"
	"cloudplane/credential-broker/internal/service"

	"github.com/gin-gonic/gin"
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
		Port:         getEnv("PORT", "8080"),
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

	// Initialize dependencies
	ctx := context.Background()

	// OIDC validator
	var oidcValidator *oidc.Validator
	if cfg.OIDCIssuer != "" {
		var err error
		oidcValidator, err = oidc.NewValidator(ctx, cfg.OIDCIssuer, cfg.OIDCAudience)
		if err != nil {
			log.Fatalf("Failed to initialize OIDC validator: %v", err)
		}
	}

	// Authorization
	authorizer := authz.NewAuthorizer()

	// AWS STS client
	stsClient, err := aws.NewSTSClient(ctx, cfg.AWSRegion)
	if err != nil {
		log.Fatalf("Failed to initialize AWS STS client: %v", err)
	}

	// Credential service
	credService := service.NewCredentialService(stsClient, cfg.MaxTTL)

	// Set up router
	router := gin.Default()
	api.SetupRouter(router, oidcValidator, authorizer, credService)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting credential broker on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
>>>>>>> Stashed changes
}
