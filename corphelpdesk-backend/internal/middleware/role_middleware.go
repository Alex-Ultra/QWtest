package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourcompany/corphelpdesk-backend/internal/models"
	"github.com/yourcompany/corphelpdesk-backend/internal/utils"
)

// RoleMiddleware checks if the user has one of the allowed roles
func RoleMiddleware(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			utils.UnauthorizedResponse(c, "User not authenticated")
			c.Abort()
			return
		}

		userModel := user.(*models.User)

		// Check if user has one of the allowed roles
		hasPermission := false
		for _, allowedRole := range allowedRoles {
			if userModel.Role == allowedRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			utils.ForbiddenResponse(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}