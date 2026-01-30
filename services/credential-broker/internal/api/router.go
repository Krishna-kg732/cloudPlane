package api

import (
	"cloudplane/credential-broker/internal/api/handlers"
	"cloudplane/credential-broker/internal/authz"
	"cloudplane/credential-broker/internal/oidc"
	"cloudplane/credential-broker/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes for the credential broker
func SetupRouter(r *gin.Engine, oidcValidator *oidc.Validator, authorizer *authz.Authorizer, credService *service.CredentialService) {
	// Create handler with dependencies
	h := handlers.NewCredentialsHandler(oidcValidator, authorizer, credService)

	// Health check (no auth)
	r.GET("/health", handlers.Health)

	// API v1 routes
	v1 := r.Group("/v1")
	{
		credentials := v1.Group("/credentials")
		{
			// Issue AWS credentials
			credentials.POST("/aws", h.IssueAWSCredentials)
		}
	}
}
