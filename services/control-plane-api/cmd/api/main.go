package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudplane/control-plane-api/internal/connections"
	"cloudplane/control-plane-api/internal/projects"

	"github.com/gin-gonic/gin"
)

// Config holds application configuration
type Config struct {
	Port                string
	TrainingServiceURL  string
	InferenceServiceURL string
}

func loadConfig() *Config {
	return &Config{
		Port:                getEnv("PORT", "8081"),
		TrainingServiceURL:  getEnv("TRAINING_SERVICE_URL", "http://localhost:8083"),
		InferenceServiceURL: getEnv("INFERENCE_SERVICE_URL", "http://localhost:8082"),
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

	// Initialize services
	projectService := projects.NewService()
	connectionService := connections.NewService()

	// Set up router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "control-plane-api"})
	})

	// API v1 routes
	v1 := router.Group("/v1")
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

		// Training jobs (proxy to training-service)
		// Route: /v1/training-jobs/* -> training-service
		v1.Any("/training-jobs", proxyHandler(cfg.TrainingServiceURL, "/v1/jobs"))
		v1.Any("/training-jobs/*path", proxyHandler(cfg.TrainingServiceURL, "/v1/jobs"))

		// Inference (proxy to inference-service)
		// Route: /v1/inference/* -> inference-service
		v1.Any("/inference", proxyHandler(cfg.InferenceServiceURL, "/v1/deployments"))
		v1.Any("/inference/*path", proxyHandler(cfg.InferenceServiceURL, "/v1/deployments"))
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

// proxyHandler creates a reverse proxy handler
// TODO: Implement actual HTTP proxying
func proxyHandler(targetURL, targetPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement reverse proxy
		// 1. Build target URL: targetURL + targetPath + c.Param("path")
		// 2. Copy request body
		// 3. Forward request with same method, headers
		// 4. Copy response back

		c.JSON(http.StatusNotImplemented, gin.H{
			"error":      "proxy not implemented",
			"target_url": targetURL,
			"path":       c.Request.URL.Path,
		})
	}
}
