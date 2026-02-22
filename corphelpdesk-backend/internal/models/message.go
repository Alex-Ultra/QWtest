package models

import (
	"time"
)

type Message struct {
	ID           string    `json:"id" db:"id"`
	TicketID     string    `json:"ticket_id" db:"ticket_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Content      string    `json:"content" db:"content"`
	Attachments  *[]string `json:"attachments,omitempty" db:"attachments"`
	IsInternal   bool      `json:"is_internal" db:"is_internal"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type CreateMessageRequest struct {
	Content     string    `json:"content" binding:"required,max=1000"`
	Attachments *[]string `json:"attachments,omitempty"`
	IsInternal  *bool     `json:"is_internal,omitempty"` // Only agents and admins can set this to true
}

type MessageResponse struct {
	ID           string    `json:"id"`
	TicketID     string    `json:"ticket_id"`
	UserID       string    `json:"user_id"`
	UserName     string    `json:"user_name"` // Name of the user who sent the message
	Content      string    `json:"content"`
	Attachments  *[]string `json:"attachments,omitempty"`
	IsInternal   bool      `json:"is_internal"`
	CreatedAt    time.Time `json:"created_at"`
}

type GetMessageListRequest struct {
	TicketID string `form:"ticket_id" binding:"required"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
}