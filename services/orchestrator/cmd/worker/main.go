package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudplane/orchestrator/internal/credclient"
	"cloudplane/orchestrator/internal/executor"
	"cloudplane/orchestrator/internal/queue"
)

// Config holds worker configuration
type Config struct {
	CredentialBrokerAddr string
	PollInterval         time.Duration
	WorkerID             string
}

func loadConfig() *Config {
	pollInterval, err := time.ParseDuration(getEnv("POLL_INTERVAL", "5s"))
	if err != nil {
		log.Printf("WARNING: Invalid POLL_INTERVAL value, using default 5s")
		pollInterval = 5 * time.Second
	}

	return &Config{
		CredentialBrokerAddr: getEnv("CREDENTIAL_BROKER_ADDR", "localhost:50051"),
		PollInterval:         pollInterval,
		WorkerID:             getEnv("WORKER_ID", "worker-1"),
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

	// Initialize gRPC client for credential broker
	credClient, err := credclient.NewClient(cfg.CredentialBrokerAddr)
	if err != nil {
		log.Fatalf("Failed to connect to credential broker: %v", err)
	}
	defer credClient.Close()

	// Initialize job queue (in-memory for MVP)
	jobQueue := queue.NewInMemoryQueue()

	// Initialize executor with gRPC credential client
	exec := executor.NewExecutor(credClient, cfg.WorkerID)

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
