package messages

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourcompany/corphelpdesk-backend/internal/models"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/auth"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/tickets"
	"github.com/yourcompany/corphelpdesk-backend/internal/utils"
)

type MessageHandler struct {
	service   tickets.TicketService
	authService auth.AuthService
}

func NewMessageHandler(service tickets.TicketService, authService auth.AuthService) *MessageHandler {
	return &MessageHandler{
		service:     service,
		authService: authService,
	}
}

// AddMessage adds a message to a ticket
// @Summary Add message to ticket
// @Description Adds a message to an existing ticket
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param message body models.CreateMessageRequest true "Message content"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 403 {object} utils.APIResponse
// @Failure 404 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /tickets/{id}/messages [post]
func (h *MessageHandler) AddMessage(c *gin.Context) {
	ticketID := c.Param("id")
	if ticketID == "" {
		utils.ValidationErrorResponse(c, "Ticket ID is required", nil)
		return
	}

	var req models.CreateMessageRequest
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

	// Check if user has permission to add message to this ticket
	ctx := context.Background()
	ticket, err := h.service.GetTicketByID(ctx, ticketID)
	if err != nil {
		utils.NotFoundResponse(c, "Ticket not found")
		return
	}

	// Allow access if user is the ticket owner, an agent, or an admin
	user, err := h.authService.GetProfile(ctx, userID.(string))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve user")
		return
	}

	if ticket.UserID != userID.(string) && user.Role != models.UserRoleAgent && user.Role != models.UserRoleAdmin && user.Role != models.UserRoleSecurityOfficer {
		utils.ForbiddenResponse(c, "You don't have permission to add messages to this ticket")
		return
	}

	// Check if user is trying to set internal message flag
	// Only agents and admins can create internal messages
	if req.IsInternal != nil && *req.IsInternal {
		if user.Role != models.UserRoleAgent && user.Role != models.UserRoleAdmin && user.Role != models.UserRoleSecurityOfficer {
			utils.ForbiddenResponse(c, "Only agents and admins can create internal messages")
			return
		}
	}

	err = h.service.AddMessage(ctx, ticketID, userID.(string), &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to add message")
		return
	}

	utils.SuccessResponse(c, map[string]string{"message": "Message added successfully"})
}

// GetMessages retrieves messages for a ticket
// @Summary Get ticket messages
// @Description Retrieves messages for a specific ticket
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.APIResponse{data=[]models.Message}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 403 {object} utils.APIResponse
// @Failure 404 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /tickets/{id}/messages [get]
func (h *MessageHandler) GetMessages(c *gin.Context) {
	ticketID := c.Param("id")
	if ticketID == "" {
		utils.ValidationErrorResponse(c, "Ticket ID is required", nil)
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page := 1
	limit := 20

	var err error
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			limit = 20
		}
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user has permission to view messages for this ticket
	ctx := context.Background()
	ticket, err := h.service.GetTicketByID(ctx, ticketID)
	if err != nil {
		utils.NotFoundResponse(c, "Ticket not found")
		return
	}

	// Allow access if user is the ticket owner, an agent, or an admin
	user, err := h.authService.GetProfile(ctx, userID.(string))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve user")
		return
	}

	if ticket.UserID != userID.(string) && user.Role != models.UserRoleAgent && user.Role != models.UserRoleAdmin && user.Role != models.UserRoleSecurityOfficer {
		utils.ForbiddenResponse(c, "You don't have permission to view messages for this ticket")
		return
	}

	messages, err := h.service.GetMessages(ctx, ticketID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve messages")
		return
	}

	utils.SuccessResponse(c, messages)
}