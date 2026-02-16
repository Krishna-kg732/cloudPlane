package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudplane/control-plane-api/internal/auth"
	"cloudplane/control-plane-api/internal/connections"
	"cloudplane/control-plane-api/internal/inferenceclient"
	"cloudplane/control-plane-api/internal/projects"
	"cloudplane/control-plane-api/internal/trainingclient"

	"github.com/gin-gonic/gin"
)

// Config holds application configuration
type Config struct {
	Port                 string
	TrainingServiceAddr  string
	InferenceServiceAddr string
}

func loadConfig() *Config {
	return &Config{
		Port:                 getEnv("PORT", "8081"),
		TrainingServiceAddr:  getEnv("TRAINING_SERVICE_ADDR", "localhost:50052"),
		InferenceServiceAddr: getEnv("INFERENCE_SERVICE_ADDR", "localhost:50053"),
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

	// Initialize gRPC clients for internal services
	trainingClient, err := trainingclient.NewClient(cfg.TrainingServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to training service: %v", err)
	}
	defer trainingClient.Close()

	inferenceClient, err := inferenceclient.NewClient(cfg.InferenceServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to inference service: %v", err)
	}
	defer inferenceClient.Close()

	// Initialize services
	projectService := projects.NewService()
	connectionService := connections.NewService()

	// Set up router
	router := gin.Default()

	// Health check (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "control-plane-api"})
	})

	// API v1 routes (JWT auth required)
	v1 := router.Group("/v1")
	v1.Use(auth.JWTAuthMiddleware())
	{
		// Projects
		v1.POST("/projects", projects.CreateHandler(projectService))
		v1.GET("/projects", projects.ListHandler(projectService))
		v1.GET("/projects/:id", projects.GetHandler(projectService))
		v1.DELETE("/projects/:id", projects.DeleteHandler(projectService))

		// Connections (nested under projects)
		v1.POST("/projects/:id/connections", connections.CreateHandler(connectionService))
		v1.GET("/projects/:id/connections", connections.ListHandler(connectionService))
		v1.DELETE("/projects/:id/connections/:connectionId", connections.DeleteHandler(connectionService))

		// Training jobs (gRPC proxy to training-service)
		v1.POST("/training-jobs", trainingJobSubmitHandler(trainingClient))
		v1.GET("/training-jobs/:id", trainingJobGetHandler(trainingClient))
		v1.GET("/training-jobs", trainingJobListHandler(trainingClient))

		// Inference (gRPC proxy to inference-service)
		v1.POST("/inference", inferenceCreateHandler(inferenceClient))
		v1.GET("/inference/:id", inferenceGetHandler(inferenceClient))
		v1.GET("/inference", inferenceListHandler(inferenceClient))
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server
	go func() {
		log.Printf("Starting control plane API on port %s", cfg.Port)
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
}

// --- Training job handlers (delegate to gRPC client) ---

func trainingJobSubmitHandler(client *trainingclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Parse request body into trainingclient.SubmitJobRequest
		// TODO: Call client.SubmitJob(c.Request.Context(), req)
		// TODO: Return job ID

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented — awaiting proto generation"})
	}
}

func trainingJobGetHandler(client *trainingclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Call client.GetJob(c.Request.Context(), c.Param("id"))
		// TODO: Return job details

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented — awaiting proto generation"})
	}
}

func trainingJobListHandler(client *trainingclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Parse query params (project_id, limit, offset)
		// TODO: Call client.ListJobs(c.Request.Context(), projectID, limit, offset)

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented — awaiting proto generation"})
	}
}

// --- Inference handlers (delegate to gRPC client) ---

func inferenceCreateHandler(client *inferenceclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Parse request body into inferenceclient.CreateDeploymentRequest
		// TODO: Call client.CreateDeployment(c.Request.Context(), req)
		// TODO: Return deployment ID

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented — awaiting proto generation"})
	}
}

func inferenceGetHandler(client *inferenceclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Call client.GetDeployment(c.Request.Context(), c.Param("id"))
		// TODO: Return deployment details

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented — awaiting proto generation"})
	}
}

func inferenceListHandler(client *inferenceclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Parse query params (project_id, limit, offset)
		// TODO: Call client.ListDeployments(c.Request.Context(), projectID, limit, offset)

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented — awaiting proto generation"})
	}
}
