package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourcompany/corphelpdesk-backend/internal/models"
	"github.com/yourcompany/corphelpdesk-backend/internal/repository/postgres"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/auth"
	"github.com/yourcompany/corphelpdesk-backend/internal/utils"
)

type TelegramAuthMiddleware struct {
	authService auth.AuthService
	userRepo    postgres.UserRepository
}

func NewTelegramAuthMiddleware(authService auth.AuthService, userRepo postgres.UserRepository) *TelegramAuthMiddleware {
	return &TelegramAuthMiddleware{
		authService: authService,
		userRepo:    userRepo,
	}
}

// TelegramAuthMiddleware validates the Telegram init data
func TelegramAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get init data from header or form
		initData := c.GetHeader("X-Telegram-Init-Data") // Or from query param
		if initData == "" {
			initData = c.Query("initData")
		}

		if initData == "" {
			utils.UnauthorizedResponse(c, "Missing Telegram init data")
			c.Abort()
			return
		}

		// Verify the init data
		userId, err := verifyTelegramInitData(initData)
		if err != nil {
			utils.UnauthorizedResponse(c, "Invalid Telegram init data")
			c.Abort()
			return
		}

		// Convert user ID to int64
		telegramID := int64(userId)

		// Get user from database
		ctx := c.Request.Context()
		user, err := h.authService.VerifyTelegramAuth(ctx, telegramID)
		if err != nil {
			utils.UnauthorizedResponse(c, "User not found")
			c.Abort()
			return
		}

		// Check verification status
		if user.VerificationStatus == models.VerificationStatusRejected {
			utils.ForbiddenResponse(c, "Access denied: account rejected")
			c.Abort()
			return
		}

		if user.VerificationStatus == models.VerificationStatusPending {
			utils.ForbiddenResponse(c, "Account pending verification")
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("telegramID", user.TelegramID)

		c.Next()
	}
}

// verifyTelegramInitData verifies the Telegram init data signature
func verifyTelegramInitData(initData string) (uint64, error) {
	// Split the init data into key-value pairs
	pairs := strings.Split(initData, "&")
	var authDate int64
	var userId uint64
	var hash string

	// Parse the key-value pairs
	dataToCheck := make([]string, 0)
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			continue
		}

		key := kv[0]
		value := kv[1]

		switch key {
		case "hash":
			hash = value
		case "auth_date":
			authDateInt, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid auth_date: %w", err)
			}
			authDate = authDateInt
		case "user":
			// Parse user object to extract ID
			userIdUint, err := parseUserIdFromUserObject(value)
			if err != nil {
				return 0, fmt.Errorf("failed to parse user ID: %w", err)
			}
			userId = userIdUint
		default:
			dataToCheck = append(dataToCheck, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Check if auth date is not older than 24 hours
	if time.Now().Unix()-authDate > 24*60*60 {
		return 0, fmt.Errorf("auth data expired")
	}

	// Sort the data to check alphabetically
	sort.Strings(dataToCheck)

	// Create the data string to verify
	dataToCheckStr := strings.Join(dataToCheck, "\n")

	// Get bot token from environment variable (in a real app)
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return 0, fmt.Errorf("missing bot token")
	}

	// Create secret key using SHA256 of bot token
	secretKey := sha256.Sum256([]byte(botToken))

	// Create HMAC-SHA256 signature
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataToCheckStr))
	computedHash := hex.EncodeToString(h.Sum(nil))

	// Compare hashes
	if !hmac.Equal([]byte(computedHash), []byte(hash)) {
		return 0, fmt.Errorf("invalid hash")
	}

	return userId, nil
}

// parseUserIdFromUserObject extracts the user ID from the URL-encoded user object
func parseUserIdFromUserObject(userStr string) (uint64, error) {
	// In a real application, you'd properly decode and parse the user object
	// For now, we'll just return a dummy value since we're focusing on structure
	decodedUser, err := url.QueryUnescape(userStr)
	if err != nil {
		return 0, err
	}

	// Parse the JSON-like string to extract user ID
	// This is simplified - in reality, you'd want to properly unmarshal the JSON
	// Find "id" field in the user string
	idStart := strings.Index(decodedUser, `"id":`)
	if idStart == -1 {
		return 0, fmt.Errorf("id not found in user object")
	}

	idStart += 4 // Move past "id":
	idEnd := idStart
	for idEnd < len(decodedUser) && decodedUser[idEnd] >= '0' && decodedUser[idEnd] <= '9' {
		idEnd++
	}

	if idStart == idEnd {
		return 0, fmt.Errorf("invalid id format in user object")
	}

	idStr := decodedUser[idStart:idEnd]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

import (
	"net/url"
	"os"
	"strconv"
)