package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourcompany/corphelpdesk-backend/internal/models"
)

type TicketRepository interface {
	Create(ctx context.Context, ticket *models.Ticket) error
	GetByID(ctx context.Context, ticketID string) (*models.Ticket, error)
	GetByUser(ctx context.Context, userID string, filter *models.TicketListFilter) ([]*models.Ticket, error)
	GetAll(ctx context.Context, filter *models.TicketListFilter) ([]*models.Ticket, error)
	UpdateStatus(ctx context.Context, ticketID string, status models.TicketStatus) error
	UpdateAssignee(ctx context.Context, ticketID string, agentID *string) error
	Update(ctx context.Context, ticket *models.Ticket) error
}

type ticketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ctx context.Context, ticket *models.Ticket) error {
	query := `
		INSERT INTO tickets (
			id, user_id, agent_id, subject, description, 
			status, priority, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	ticket.ID = uuid.New().String()
	ticket.CreatedAt = now
	ticket.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		ticket.ID, ticket.UserID, ticket.AgentID, ticket.Subject,
		ticket.Description, ticket.Status, ticket.Priority,
		ticket.CreatedAt, ticket.UpdatedAt)

	return err
}

func (r *ticketRepository) GetByID(ctx context.Context, ticketID string) (*models.Ticket, error) {
	query := `
		SELECT id, user_id, agent_id, subject, description, 
		       status, priority, created_at, updated_at
		FROM tickets WHERE id = $1
	`

	var ticket models.Ticket
	err := r.db.QueryRowContext(ctx, query, ticketID).Scan(
		&ticket.ID, &ticket.UserID, &ticket.AgentID, &ticket.Subject,
		&ticket.Description, &ticket.Status, &ticket.Priority,
		&ticket.CreatedAt, &ticket.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ticket with id %s not found", ticketID)
		}
		return nil, err
	}

	return &ticket, nil
}

func (r *ticketRepository) GetByUser(ctx context.Context, userID string, filter *models.TicketListFilter) ([]*models.Ticket, error) {
	query := `SELECT id, user_id, agent_id, subject, description, status, priority, created_at, updated_at FROM tickets WHERE user_id = $1`
	args := []interface{}{userID}
	argCount := 2

	// Add filters if present
	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filter.Status)
		argCount++
	}

	if filter.Priority != nil {
		query += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, *filter.Priority)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}
	if filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		err := rows.Scan(
			&ticket.ID, &ticket.UserID, &ticket.AgentID, &ticket.Subject,
			&ticket.Description, &ticket.Status, &ticket.Priority,
			&ticket.CreatedAt, &ticket.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &ticket)
	}

	return tickets, nil
}

func (r *ticketRepository) GetAll(ctx context.Context, filter *models.TicketListFilter) ([]*models.Ticket, error) {
	query := `SELECT id, user_id, agent_id, subject, description, status, priority, created_at, updated_at FROM tickets`
	args := []interface{}{}
	argCount := 1

	// Add filters if present
	whereAdded := false
	if filter.Status != nil {
		query += fmt.Sprintf(" WHERE status = $%d", argCount)
		args = append(args, *filter.Status)
		argCount++
		whereAdded = true
	}

	if filter.Priority != nil {
		if !whereAdded {
			query += fmt.Sprintf(" WHERE priority = $%d", argCount)
			whereAdded = true
		} else {
			query += fmt.Sprintf(" AND priority = $%d", argCount)
		}
		args = append(args, *filter.Priority)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}
	if filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		err := rows.Scan(
			&ticket.ID, &ticket.UserID, &ticket.AgentID, &ticket.Subject,
			&ticket.Description, &ticket.Status, &ticket.Priority,
			&ticket.CreatedAt, &ticket.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &ticket)
	}

	return tickets, nil
}

func (r *ticketRepository) UpdateStatus(ctx context.Context, ticketID string, status models.TicketStatus) error {
	query := `
		UPDATE tickets 
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, status, now, ticketID)
	return err
}

func (r *ticketRepository) UpdateAssignee(ctx context.Context, ticketID string, agentID *string) error {
	query := `
		UPDATE tickets 
		SET agent_id = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, agentID, now, ticketID)
	return err
}

func (r *ticketRepository) Update(ctx context.Context, ticket *models.Ticket) error {
	query := `
		UPDATE tickets 
		SET agent_id = $1, subject = $2, description = $3, 
		    status = $4, priority = $5, updated_at = $6
		WHERE id = $7
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		ticket.AgentID, ticket.Subject, ticket.Description,
		ticket.Status, ticket.Priority, now, ticket.ID)

	return err
}