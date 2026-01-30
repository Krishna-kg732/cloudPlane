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
	port := getEnv("PORT", "8083")

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "training-service"})
	})

	// API v1
	v1 := router.Group("/v1")
	{
		// Training Jobs
		v1.POST("/jobs", submitJobHandler)
		v1.GET("/jobs", listJobsHandler)
		v1.GET("/jobs/:id", getJobHandler)
		v1.POST("/jobs/:id/cancel", cancelJobHandler)
		v1.GET("/jobs/:id/logs", getJobLogsHandler)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting training service on port %s", port)
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

func submitJobHandler(c *gin.Context) {
	// TODO: Implementation should:
	// 1. Parse request: framework, image, workers, gpus_per_worker, storage config
	// 2. Validate project has a cloud connection
	// 3. Create job record in DB with status=pending
	// 4. Queue job for orchestrator
	// 5. Return job object

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func listJobsHandler(c *gin.Context) {
	// TODO: List jobs, filter by project_id and status
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func getJobHandler(c *gin.Context) {
	// TODO: Get job by ID, include status, cost, duration
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func cancelJobHandler(c *gin.Context) {
	// TODO: Cancel running/pending job
	// 1. Update job status to cancelled
	// 2. Notify orchestrator to stop execution
	// 3. Cleanup K8s resources if running
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func getJobLogsHandler(c *gin.Context) {
	// TODO: Stream logs from K8s pods
	// 1. Get job's K8s namespace and pod names
	// 2. Stream logs via kubectl logs or K8s API
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
