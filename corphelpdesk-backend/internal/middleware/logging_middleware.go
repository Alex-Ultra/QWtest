package middleware

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourcompany/corphelpdesk-backend/internal/repository/postgres"
)

// LoggingMiddleware logs requests to the access_logs table
func LoggingMiddleware(accessLogRepo postgres.AccessLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read request body if needed for logging
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
			// Restore the io.ReadCloser to its original state
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get user ID if available
		userID, exists := c.Get("userID")
		var userIDStr *string
		if exists {
			id := userID.(string)
			userIDStr = &id
		}

		// Log entry
		logEntry := &AccessLog{
			UserID:    userIDStr,
			Action:    c.Request.Method + " " + c.Request.URL.Path,
			IPAddress: c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Timestamp: start,
		}

		// Store in database asynchronously to avoid blocking the request
		go func() {
			// In a real implementation, you would save the log to the database
			// accessLogRepo.Create(context.Background(), logEntry)
		}()
	}
}

// AccessLog represents a log entry for user actions
type AccessLog struct {
	ID        string    `json:"id" db:"id"`
	UserID    *string   `json:"user_id,omitempty" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// AccessLogRepository defines methods for accessing log data
type AccessLogRepository interface {
	Create(ctx context.Context, log *AccessLog) error
	GetByUser(ctx context.Context, userID string, limit, offset int) ([]*AccessLog, error)
	GetAll(ctx context.Context, limit, offset int) ([]*AccessLog, error)
}

import (
	"context"
)