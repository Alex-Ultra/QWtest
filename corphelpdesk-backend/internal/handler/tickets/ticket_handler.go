package tickets

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourcompany/corphelpdesk-backend/internal/models"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/auth"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/tickets"
	"github.com/yourcompany/corphelpdesk-backend/internal/utils"
)

type TicketHandler struct {
	service   tickets.TicketService
	authService auth.AuthService
}

func NewTicketHandler(service tickets.TicketService, authService auth.AuthService) *TicketHandler {
	return &TicketHandler{
		service:     service,
		authService: authService,
	}
}

// GetTickets retrieves tickets for the authenticated user
// @Summary Get user tickets
// @Description Retrieves a list of tickets created by the authenticated user
// @Tags tickets
// @Accept json
// @Produce json
// @Param status query string false "Ticket status filter"
// @Param priority query string false "Ticket priority filter"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.APIResponse{data=[]models.Ticket}
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /tickets [get]
func (h *TicketHandler) GetTickets(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var filter models.TicketListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.ValidationErrorResponse(c, "Invalid query parameters", map[string]interface{}{"error": err.Error()})
		return
	}

	ctx := context.Background()
	tickets, err := h.service.GetTicketsByUser(ctx, userID.(string), &filter)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve tickets")
		return
	}

	utils.SuccessResponse(c, tickets)
}

// CreateTicket creates a new ticket
// @Summary Create a new ticket
// @Description Creates a new support ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Param ticket body models.CreateTicketRequest true "Ticket details"
// @Success 201 {object} utils.APIResponse{data=models.Ticket}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /tickets [post]
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var req models.CreateTicketRequest
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

	ctx := context.Background()
	ticket, err := h.service.CreateTicket(ctx, userID.(string), &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create ticket")
		return
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Status:    utils.ResponseStatusSuccess,
		Data:      ticket,
		Error:     nil,
		Timestamp: ticket.CreatedAt,
	})
}

// GetTicketByID retrieves a specific ticket by ID
// @Summary Get ticket by ID
// @Description Retrieves a ticket by its ID along with its messages
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Success 200 {object} utils.APIResponse{data=models.Ticket}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 404 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /tickets/{id} [get]
func (h *TicketHandler) GetTicketByID(c *gin.Context) {
	ticketID := c.Param("id")
	if ticketID == "" {
		utils.ValidationErrorResponse(c, "Ticket ID is required", nil)
		return
	}

	ctx := context.Background()
	ticket, err := h.service.GetTicketByID(ctx, ticketID)
	if err != nil {
		utils.NotFoundResponse(c, "Ticket not found")
		return
	}

	// Check if user has permission to view this ticket
	userID, exists := c.Get("userID")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Allow access if user is the ticket owner, an agent, or an admin
	user, err := h.authService.GetProfile(ctx, userID.(string))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve user")
		return
	}

	if ticket.UserID != userID.(string) && user.Role != models.UserRoleAgent && user.Role != models.UserRoleAdmin && user.Role != models.UserRoleSecurityOfficer {
		utils.ForbiddenResponse(c, "You don't have permission to view this ticket")
		return
	}

	utils.SuccessResponse(c, ticket)
}

// UpdateTicketStatus updates the status of a ticket
// @Summary Update ticket status
// @Description Updates the status of a ticket (agents and admins only)
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param status body models.UpdateTicketStatusRequest true "New status"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 403 {object} utils.APIResponse
// @Failure 404 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /tickets/{id}/status [patch]
func (h *TicketHandler) UpdateTicketStatus(c *gin.Context) {
	ticketID := c.Param("id")
	if ticketID == "" {
		utils.ValidationErrorResponse(c, "Ticket ID is required", nil)
		return
	}

	var req models.UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, "Invalid request data", map[string]interface{}{"error": err.Error()})
		return
	}

	ctx := context.Background()
	err := h.service.UpdateTicketStatus(ctx, ticketID, req.Status)
	if err != nil {
		utils.NotFoundResponse(c, "Ticket not found")
		return
	}

	utils.SuccessResponse(c, map[string]string{"message": "Ticket status updated successfully"})
}

// AssignTicket assigns a ticket to an agent
// @Summary Assign ticket to agent
// @Description Assigns a ticket to an agent (admins only)
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param assignment body models.AssignTicketRequest true "Assignment details"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 403 {object} utils.APIResponse
// @Failure 404 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /tickets/{id}/assign [patch]
func (h *TicketHandler) AssignTicket(c *gin.Context) {
	ticketID := c.Param("id")
	if ticketID == "" {
		utils.ValidationErrorResponse(c, "Ticket ID is required", nil)
		return
	}

	var req models.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, "Invalid request data", map[string]interface{}{"error": err.Error()})
		return
	}

	ctx := context.Background()
	err := h.service.AssignTicket(ctx, ticketID, req.AgentID)
	if err != nil {
		utils.NotFoundResponse(c, "Ticket not found")
		return
	}

	utils.SuccessResponse(c, map[string]string{"message": "Ticket assigned successfully"})
}