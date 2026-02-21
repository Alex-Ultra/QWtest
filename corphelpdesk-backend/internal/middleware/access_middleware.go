package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourcompany/corphelpdesk-backend/internal/models"
	"github.com/yourcompany/corphelpdesk-backend/internal/utils"
)

// AccessMiddleware checks if the user's verification status allows access
func AccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			utils.UnauthorizedResponse(c, "User not authenticated")
			c.Abort()
			return
		}

		userModel := user.(*models.User)

		// Check verification status
		switch userModel.VerificationStatus {
		case models.VerificationStatusRejected:
			utils.ForbiddenResponse(c, "Access denied: account rejected")
			c.Abort()
			return
		case models.VerificationStatusPending:
			utils.ForbiddenResponse(c, "Account pending verification")
			c.Abort()
			return
		case models.VerificationStatusVerified:
			// User is verified, allow access
			c.Next()
			return
		default:
			utils.ForbiddenResponse(c, "Invalid verification status")
			c.Abort()
			return
		}
	}
}