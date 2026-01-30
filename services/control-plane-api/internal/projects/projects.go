package projects

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Project represents a cloudplane project
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Repository interface for project storage
// TODO: Implement PostgreSQL repository
type Repository interface {
	Create(name string) (*Project, error)
	Get(id string) (*Project, error)
	List() ([]*Project, error)
	Delete(id string) error
}

// Service handles project operations
type Service struct {
	repo Repository
}

// NewService creates a new project service
func NewService() *Service {
	// TODO: Replace with PostgreSQL repository
	// repo := NewPostgresRepository(db)
	return &Service{
		repo: nil, // Placeholder - implement repository
	}
}

// --- HTTP Handlers ---

type createRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateHandler handles project creation
func CreateHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Call svc.repo.Create(req.Name)
		// Return created project
		_ = req

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
	}
}

// GetHandler retrieves a project
func GetHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// TODO: Call svc.repo.Get(id)
		// Return project or 404
		_ = id

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
	}
}

// ListHandler lists all projects
func ListHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Call svc.repo.List()
		// Return projects array

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
	}
}

// DeleteHandler deletes a project
func DeleteHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// TODO: Call svc.repo.Delete(id)
		// Return success or 404
		_ = id

		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
	}
}
