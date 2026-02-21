package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourcompany/corphelpdesk-backend/internal/utils"
)

type RateLimiter struct {
	visitors map[string]*Visitor
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

type Visitor struct {
	Hits  int
	Reset time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine to remove old entries
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists || now.After(v.Reset) {
		rl.visitors[ip] = &Visitor{
			Hits:  1,
			Reset: now.Add(rl.window),
		}
		return true
	}

	if v.Hits < rl.limit {
		v.Hits++
		return true
	}

	return false
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.After(v.Reset) {
				delete(rl.visitors, ip)
			}
		}
		rl.mutex.Unlock()
	}
}

var (
	// Different rate limits for different endpoints
	normalRateLimiter   = NewRateLimiter(100, time.Minute)   // 100 requests per minute for normal users
	verificationRateLimiter = NewRateLimiter(10, time.Minute)    // 10 requests per minute for verification endpoints
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Determine which rate limiter to use based on endpoint
		var limiter *RateLimiter
		if c.Request.URL.Path == "/api/auth/verify" || c.Request.URL.Path == "/api/users/profile" {
			limiter = verificationRateLimiter
		} else {
			limiter = normalRateLimiter
		}

		if !limiter.Allow(ip) {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded", map[string]interface{}{
				"retry_after": "60",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}