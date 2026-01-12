package auth

import (
	"context"
	"net/http"
	"strings"
	"pr-server/internal/config"
)

type contextKey string

const AgentIDKey contextKey = "agent_id"

// AuthMiddleware validates the token in the Authorization header
// Skip authentication for static files and web interface routes
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for static files and web interface routes
			if r.URL.Path == "/" || 
				strings.HasPrefix(r.URL.Path, "/styles.css") ||
				strings.HasPrefix(r.URL.Path, "/app.js") ||
				strings.HasPrefix(r.URL.Path, "/favicon.ico") ||
				strings.HasPrefix(r.URL.Path, "/assets/") {
				next.ServeHTTP(w, r)
				return
			}

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