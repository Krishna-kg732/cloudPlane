package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudplane/orchestrator/internal/executor"
	"cloudplane/orchestrator/internal/queue"
)

// Config holds worker configuration
type Config struct {
	CredentialBrokerURL string
	PollInterval        time.Duration
	WorkerID            string
}

func loadConfig() *Config {
	return &Config{
		CredentialBrokerURL: getEnv("CREDENTIAL_BROKER_URL", "http://localhost:8080"),
		PollInterval:        5 * time.Second,
		WorkerID:            getEnv("WORKER_ID", "worker-1"),
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize job queue (in-memory for MVP)
	jobQueue := queue.NewInMemoryQueue()

	// Initialize executor
	exec := executor.NewExecutor(cfg.CredentialBrokerURL, cfg.WorkerID)

	// Start worker loop
	go func() {
		log.Printf("Starting orchestrator worker %s", cfg.WorkerID)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Poll for jobs
				job, err := jobQueue.Poll()
				if err != nil {
					log.Printf("Error polling queue: %v", err)
					time.Sleep(cfg.PollInterval)
					continue
				}

				if job != nil {
					log.Printf("Processing job: %s", job.ID)
					if err := exec.Execute(ctx, job); err != nil {
						log.Printf("Job %s failed: %v", job.ID, err)
						if markErr := jobQueue.MarkFailed(job.ID, err.Error()); markErr != nil {
							log.Printf("Failed to mark job failed: %v", markErr)
						}
					} else {
						log.Printf("Job %s completed", job.ID)
						if markErr := jobQueue.MarkCompleted(job.ID); markErr != nil {
							log.Printf("Failed to mark job completed: %v", markErr)
						}
					}
				} else {
					time.Sleep(cfg.PollInterval)
				}
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down worker...")
	cancel()
	log.Println("Worker exited")
}
