package training

// NOTE: Training job logic has been moved to training-service
// This package is kept for backwards compatibility during migration
//
// See: services/training-service/
//
// The control-plane-api now only routes training requests to training-service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProxyHandler forwards training requests to training-service
// TODO: Implement HTTP proxy to training-service
func ProxyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Forward request to training-service
		// trainingServiceURL := os.Getenv("TRAINING_SERVICE_URL")
		// proxy.Forward(c, trainingServiceURL)

		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "training-service proxy not implemented",
			"message": "Training logic moved to training-service",
		})
	}
}
