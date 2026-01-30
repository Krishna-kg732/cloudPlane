package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Middleware provides authentication middleware for the API
// TODO: Implement JWT validation and API key auth

// JWTAuthMiddleware validates JWT tokens
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement JWT validation
		//
		// 1. Extract token from Authorization header
		// 2. Validate token signature and expiration
		// 3. Extract user claims (sub, email, project_id)
		// 4. Set claims in context for handlers
		//
		// Example:
		// authHeader := c.GetHeader("Authorization")
		// if authHeader == "" {
		//     c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization"})
		//     return
		// }
		// token := strings.TrimPrefix(authHeader, "Bearer ")
		// claims, err := validateToken(token)
		// if err != nil {
		//     c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
		//     return
		// }
		// c.Set("user_id", claims.Subject)
		// c.Set("project_id", claims.ProjectID)

		c.Next()
	}
}

// RequireProject middleware ensures request has valid project access
func RequireProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Validate user has access to the requested project
		//
		// projectID := c.Param("id")
		// userID := c.GetString("user_id")
		// if !hasProjectAccess(userID, projectID) {
		//     c.AbortWithStatusJSON(403, gin.H{"error": "project access denied"})
		//     return
		// }

		c.Next()
	}
}

// HealthHandler returns service health (no auth required)
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
