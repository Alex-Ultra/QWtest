package models

import (
	"time"
)

type UserRole string
type VerificationStatus string

const (
	UserRoleClient         UserRole = "client"
	UserRoleAgent          UserRole = "agent"
	UserRoleAdmin          UserRole = "admin"
	UserRoleSecurityOfficer UserRole = "security_officer"
)

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusRejected VerificationStatus = "rejected"
)

type User struct {
	ID                  string               `json:"id" db:"id"`
	TelegramID          int64                `json:"telegram_id" db:"telegram_id"`
	PhoneEncrypted      []byte               `json:"-" db:"phone_enc"`
	SNILSEncrypted      []byte               `json:"-" db:"snils_enc"`
	DOBEncrypted        []byte               `json:"-" db:"dob_enc"`
	FullName            string               `json:"full_name" db:"full_name"`
	Username            *string              `json:"username,omitempty" db:"username"`
	Role                UserRole             `json:"role" db:"role"`
	VerificationStatus  VerificationStatus   `json:"verification_status" db:"verification_status"`
	PDConsentAccepted   bool                 `json:"pd_consent_accepted" db:"pd_consent_accepted"`
	RejectionReason     *string              `json:"rejection_reason,omitempty" db:"rejection_reason"`
	CreatedAt           time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at" db:"updated_at"`
}

type UserProfileUpdate struct {
	SNILS       string    `json:"snils" binding:"required"`
	DateOfBirth time.Time `json:"date_of_birth" binding:"required"`
	Phone       string    `json:"phone" binding:"required"`
	PDConsent   bool      `json:"pd_consent" binding:"required"`
}

type UserProfileResponse struct {
	ID                  string               `json:"id"`
	TelegramID          int64                `json:"telegram_id"`
	FullName            string               `json:"full_name"`
	Username            *string              `json:"username,omitempty"`
	Role                UserRole             `json:"role"`
	VerificationStatus  VerificationStatus   `json:"verification_status"`
	RejectionReason     *string              `json:"rejection_reason,omitempty"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

// For security officers who need to see decrypted data
type UserProfileWithPersonalData struct {
	ID                  string               `json:"id"`
	TelegramID          int64                `json:"telegram_id"`
	FullName            string               `json:"full_name"`
	Username            *string              `json:"username,omitempty"`
	Role                UserRole             `json:"role"`
	VerificationStatus  VerificationStatus   `json:"verification_status"`
	SNILS               string               `json:"snils"`
	DateOfBirth         string               `json:"date_of_birth"` // Format: YYYY-MM-DD
	Phone               string               `json:"phone"`
	RejectionReason     *string              `json:"rejection_reason,omitempty"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}