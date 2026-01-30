package connections

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Connection represents a cloud account connection
type Connection struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Provider  string    `json:"provider"` // aws, gcp, azure
	RoleARN   string    `json:"role_arn"` // AWS IAM role ARN (or equivalent)
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository interface for connection storage
// TODO: Implement PostgreSQL repository
type Repository interface {
	Create(projectID, provider, roleARN, region string) (*Connection, error)
	Get(id string) (*Connection, error)
	GetByProject(projectID string) ([]*Connection, error)
	Delete(id string) error
}

// Service handles connection operations
type Service struct {
	repo Repository
}

// NewService creates a new connection service
func NewService() *Service {
	// TODO: Replace with PostgreSQL repository
	return &Service{
		repo: nil,
	}
}

// --- HTTP Handlers ---

type createRequest struct {
	Provider string `json:"provider" binding:"required"`
	RoleARN  string `json:"role_arn" binding:"required"`
	Region   string `json:"region" binding:"required"`
}

// CreateHandler handles connection creation
func CreateHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Validate provider (aws, gcp, azure)
		// TODO: Validate role_arn format based on provider
		// TODO: Call svc.repo.Create(projectID, req.Provider, req.RoleARN, req.Region)
		_ = projectID
		_ = req

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
	}
}

// ListHandler lists connections for a project
func ListHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		// TODO: Call svc.repo.GetByProject(projectID)
		_ = projectID

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
	}
}

// DeleteHandler deletes a connection
func DeleteHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		connectionID := c.Param("connectionId")

		// TODO: Call svc.repo.Delete(connectionID)
		_ = connectionID

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
	}
}
