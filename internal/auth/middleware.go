package auth

import (
	"context"
	"net/http"
	"strings"
	"monitor-server/internal/config"
)

type contextKey string

const AgentIDKey contextKey = "agent_id"

// AuthMiddleware validates the token in the Authorization header
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			token = strings.TrimSpace(token)

			if agentID, exists := cfg.Tokens[token]; exists {
				// Add agent ID to context
				ctx := context.WithValue(r.Context(), AgentIDKey, agentID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			}
		})
	}
}