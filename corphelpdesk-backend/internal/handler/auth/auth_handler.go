package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourcompany/corphelpdesk-backend/internal/models"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/auth"
	"github.com/yourcompany/corphelpdesk-backend/internal/utils"
)

type AuthHandler struct {
	service auth.AuthService
}

func NewAuthHandler(service auth.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// VerifyTelegramAuth verifies the Telegram authentication data
// @Summary Verify Telegram authentication
// @Description Verifies the Telegram user authentication and creates/returns user info
// @Tags auth
// @Accept json
// @Produce json
// @Param initData body map[string]string true "Telegram init data"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /auth/verify [post]
func (h *AuthHandler) VerifyTelegramAuth(c *gin.Context) {
	// Get user from context (set by Telegram auth middleware)
	user, exists := c.Get("user")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	telegramUser := user.(*models.User)
	utils.SuccessResponse(c, telegramUser)
}

// GetProfile retrieves the authenticated user's profile
// @Summary Get user profile
// @Description Retrieves the profile of the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} utils.APIResponse{data=models.UserProfileResponse}
// @Failure 401 {object} utils.APIResponse
// @Failure 403 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /users/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	ctx := context.Background()
	user, err := h.service.GetProfile(ctx, userID.(string))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve profile")
		return
	}

	response := models.UserProfileResponse{
		ID:                 user.ID,
		TelegramID:         user.TelegramID,
		FullName:           user.FullName,
		Username:           user.Username,
		Role:               user.Role,
		VerificationStatus: user.VerificationStatus,
		RejectionReason:    user.RejectionReason,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}

	utils.SuccessResponse(c, response)
}

// SubmitProfile submits user profile information for verification
// @Summary Submit user profile
// @Description Submits user profile information including SNILS and date of birth for verification
// @Tags auth
// @Accept json
// @Produce json
// @Param profile body models.UserProfileUpdate true "User profile information"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 403 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /users/profile [post]
func (h *AuthHandler) SubmitProfile(c *gin.Context) {
	var req models.UserProfileUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, "Invalid request data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user has already submitted profile
	ctx := context.Background()
	user, err := h.service.GetProfile(ctx, userID.(string))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve user")
		return
	}

	if user.VerificationStatus != models.VerificationStatusPending {
		utils.ForbiddenResponse(c, "Profile can only be submitted when verification status is pending")
		return
	}

	// Update profile
	err = h.service.UpdateProfile(ctx, userID.(string), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "PROFILE_UPDATE_FAILED", err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, map[string]string{"message": "Profile submitted successfully"})
}