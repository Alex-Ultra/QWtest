package models

import (
	"time"
)

type TicketStatus string
type TicketPriority string

const (
	TicketStatusNew      TicketStatus = "new"
	TicketStatusOpen     TicketStatus = "open"
	TicketStatusPending  TicketStatus = "pending"
	TicketStatusResolved TicketStatus = "resolved"
	TicketStatusClosed   TicketStatus = "closed"
)

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityNormal TicketPriority = "normal"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityCritical TicketPriority = "critical"
)

type Ticket struct {
	ID          string         `json:"id" db:"id"`
	UserID      string         `json:"user_id" db:"user_id"`
	AgentID     *string        `json:"agent_id,omitempty" db:"agent_id"`
	Subject     string         `json:"subject" db:"subject"`
	Description string         `json:"description" db:"description"`
	Status      TicketStatus   `json:"status" db:"status"`
	Priority    TicketPriority `json:"priority" db:"priority"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

type CreateTicketRequest struct {
	Subject     string         `json:"subject" binding:"required,max=200"`
	Description string         `json:"description" binding:"required"`
	Priority    TicketPriority `json:"priority" binding:"required,oneof=low normal high critical"`
}

type UpdateTicketStatusRequest struct {
	Status TicketStatus `json:"status" binding:"required,oneof=new open pending resolved closed"`
}

type AssignTicketRequest struct {
	AgentID string `json:"agent_id" binding:"required"`
}

type TicketResponse struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	AgentID     *string        `json:"agent_id,omitempty"`
	Subject     string         `json:"subject"`
	Description string         `json:"description"`
	Status      TicketStatus   `json:"status"`
	Priority    TicketPriority `json:"priority"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	UserName    string         `json:"user_name"` // Name of the user who created the ticket
	AgentName   *string        `json:"agent_name,omitempty"` // Name of the assigned agent
}

type TicketListFilter struct {
	Status   *TicketStatus    `form:"status"`
	Priority *TicketPriority  `form:"priority"`
	Page     int              `form:"page,default=1"`
	Limit    int              `form:"limit,default=10"`
}