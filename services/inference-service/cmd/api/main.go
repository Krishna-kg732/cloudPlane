package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	port := getEnv("PORT", "8082")

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "inference-service"})
	})

	// API v1
	v1 := router.Group("/v1")
	{
		// Deployments
		v1.POST("/deployments", createDeploymentHandler)
		v1.GET("/deployments", listDeploymentsHandler)
		v1.GET("/deployments/:id", getDeploymentHandler)
		v1.DELETE("/deployments/:id", deleteDeploymentHandler)

		// Inference (proxy to deployed model)
		v1.POST("/inference/:deployment_id", inferenceHandler)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting inference service on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// --- Handlers (TODO stubs) ---

func createDeploymentHandler(c *gin.Context) {
	// TODO: Implementation should:
	// 1. Parse request: model, engine (vllm/tgi/triton), instance_type, replicas
	// 2. Validate model exists (HuggingFace or S3 path)
	// 3. Create deployment record in DB
	// 4. Queue deployment job for orchestrator
	// 5. Return deployment object with status=pending

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func listDeploymentsHandler(c *gin.Context) {
	// TODO: List deployments, optionally filter by project_id
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func getDeploymentHandler(c *gin.Context) {
	// TODO: Get deployment by ID, include status and endpoint URL
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func deleteDeploymentHandler(c *gin.Context) {
	// TODO: Delete deployment, cleanup K8s resources
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func inferenceHandler(c *gin.Context) {
	// TODO: Proxy inference request to deployed model
	// 1. Get deployment endpoint from DB
	// 2. Forward request to model endpoint
	// 3. Track token usage for cost calculation
	// 4. Return model response

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
