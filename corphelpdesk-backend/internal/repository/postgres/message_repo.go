package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourcompany/corphelpdesk-backend/internal/models"
)

type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	GetByTicketID(ctx context.Context, ticketID string, page, limit int) ([]*models.Message, error)
	GetByID(ctx context.Context, messageID string) (*models.Message, error)
}

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (
			id, ticket_id, user_id, content, attachments, is_internal, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	message.ID = uuid.New().String()
	message.CreatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		message.ID, message.TicketID, message.UserID, message.Content,
		message.Attachments, message.IsInternal, message.CreatedAt)

	return err
}

func (r *messageRepository) GetByTicketID(ctx context.Context, ticketID string, page, limit int) ([]*models.Message, error) {
	query := `
		SELECT id, ticket_id, user_id, content, attachments, is_internal, created_at
		FROM messages 
		WHERE ticket_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx, query, ticketID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var message models.Message
		err := rows.Scan(
			&message.ID, &message.TicketID, &message.UserID, &message.Content,
			&message.Attachments, &message.IsInternal, &message.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

func (r *messageRepository) GetByID(ctx context.Context, messageID string) (*models.Message, error) {
	query := `
		SELECT id, ticket_id, user_id, content, attachments, is_internal, created_at
		FROM messages 
		WHERE id = $1
	`

	var message models.Message
	err := r.db.QueryRowContext(ctx, query, messageID).Scan(
		&message.ID, &message.TicketID, &message.UserID, &message.Content,
		&message.Attachments, &message.IsInternal, &message.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message with id %s not found", messageID)
		}
		return nil, err
	}

	return &message, nil
}