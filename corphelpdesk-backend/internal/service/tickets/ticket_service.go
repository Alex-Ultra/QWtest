package tickets

import (
	"context"
	"fmt"

	"github.com/yourcompany/corphelpdesk-backend/internal/models"
	"github.com/yourcompany/corphelpdesk-backend/internal/repository/postgres"
)

type TicketService interface {
	CreateTicket(ctx context.Context, userID string, req *models.CreateTicketRequest) (*models.Ticket, error)
	GetTicketByID(ctx context.Context, ticketID string) (*models.Ticket, error)
	GetTicketsByUser(ctx context.Context, userID string, filter *models.TicketListFilter) ([]*models.Ticket, error)
	GetAllTickets(ctx context.Context, filter *models.TicketListFilter) ([]*models.Ticket, error)
	UpdateTicketStatus(ctx context.Context, ticketID string, status models.TicketStatus) error
	AssignTicket(ctx context.Context, ticketID string, agentID string) error
	AddMessage(ctx context.Context, ticketID, userID string, req *models.CreateMessageRequest) error
	GetMessages(ctx context.Context, ticketID string, page, limit int) ([]*models.Message, error)
}

type ticketService struct {
	ticketRepo  postgres.TicketRepository
	messageRepo postgres.MessageRepository
}

func NewTicketService(ticketRepo postgres.TicketRepository, messageRepo postgres.MessageRepository) TicketService {
	return &ticketService{
		ticketRepo:  ticketRepo,
		messageRepo: messageRepo,
	}
}

func (s *ticketService) CreateTicket(ctx context.Context, userID string, req *models.CreateTicketRequest) (*models.Ticket, error) {
	ticket := &models.Ticket{
		UserID:      userID,
		Subject:     req.Subject,
		Description: req.Description,
		Status:      models.TicketStatusNew,
		Priority:    req.Priority,
	}

	if err := s.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// Create initial message with the ticket description
	initialMessage := &models.Message{
		TicketID:   ticket.ID,
		UserID:     userID,
		Content:    req.Description,
		IsInternal: false,
	}

	if err := s.messageRepo.Create(ctx, initialMessage); err != nil {
		return nil, fmt.Errorf("failed to create initial message: %w", err)
	}

	return ticket, nil
}

func (s *ticketService) GetTicketByID(ctx context.Context, ticketID string) (*models.Ticket, error) {
	return s.ticketRepo.GetByID(ctx, ticketID)
}

func (s *ticketService) GetTicketsByUser(ctx context.Context, userID string, filter *models.TicketListFilter) ([]*models.Ticket, error) {
	return s.ticketRepo.GetByUser(ctx, userID, filter)
}

func (s *ticketService) GetAllTickets(ctx context.Context, filter *models.TicketListFilter) ([]*models.Ticket, error) {
	return s.ticketRepo.GetAll(ctx, filter)
}

func (s *ticketService) UpdateTicketStatus(ctx context.Context, ticketID string, status models.TicketStatus) error {
	return s.ticketRepo.UpdateStatus(ctx, ticketID, status)
}

func (s *ticketService) AssignTicket(ctx context.Context, ticketID string, agentID string) error {
	return s.ticketRepo.UpdateAssignee(ctx, ticketID, &agentID)
}

func (s *ticketService) AddMessage(ctx context.Context, ticketID, userID string, req *models.CreateMessageRequest) error {
	message := &models.Message{
		TicketID:    ticketID,
		UserID:      userID,
		Content:     req.Content,
		Attachments: req.Attachments,
		IsInternal:  req.IsInternal != nil && *req.IsInternal,
	}

	return s.messageRepo.Create(ctx, message)
}

func (s *ticketService) GetMessages(ctx context.Context, ticketID string, page, limit int) ([]*models.Message, error) {
	return s.messageRepo.GetByTicketID(ctx, ticketID, page, limit)
}