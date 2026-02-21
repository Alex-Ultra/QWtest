package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourcompany/corphelpdesk-backend/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	UpdateVerificationStatus(ctx context.Context, userID string, status models.VerificationStatus, reason *string) error
	UpdateUserProfile(ctx context.Context, userID string, snilsEncrypted, phoneEncrypted, dobEncrypted []byte, consent bool) error
	FindPendingVerifications(ctx context.Context) ([]*models.User, error)
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetWithDecryptedData(ctx context.Context, userID string) (*models.UserProfileWithPersonalData, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.User, error)
	RevokeAccess(ctx context.Context, userID string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, telegram_id, phone_enc, snils_enc, dob_enc, full_name, 
			username, role, verification_status, pd_consent_accepted, 
			rejection_reason, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	now := time.Now()
	user.ID = uuid.New().String()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.TelegramID, user.PhoneEncrypted, user.SNILSEncrypted,
		user.DOBEncrypted, user.FullName, user.Username, user.Role,
		user.VerificationStatus, user.PDConsentAccepted, user.RejectionReason,
		user.CreatedAt, user.UpdatedAt)

	return err
}

func (r *userRepository) FindByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, phone_enc, snils_enc, dob_enc, full_name, 
		       username, role, verification_status, pd_consent_accepted, 
		       rejection_reason, created_at, updated_at
		FROM users WHERE telegram_id = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.PhoneEncrypted, &user.SNILSEncrypted,
		&user.DOBEncrypted, &user.FullName, &user.Username, &user.Role,
		&user.VerificationStatus, &user.PDConsentAccepted, &user.RejectionReason,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with telegram_id %d not found", telegramID)
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateVerificationStatus(ctx context.Context, userID string, status models.VerificationStatus, reason *string) error {
	query := `
		UPDATE users 
		SET verification_status = $1, rejection_reason = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, status, reason, now, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) UpdateUserProfile(ctx context.Context, userID string, snilsEncrypted, phoneEncrypted, dobEncrypted []byte, consent bool) error {
	query := `
		UPDATE users 
		SET snils_enc = $1, phone_enc = $2, dob_enc = $3, 
		    pd_consent_accepted = $4, verification_status = $5, updated_at = $6
		WHERE id = $7
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		snilsEncrypted, phoneEncrypted, dobEncrypted,
		consent, models.VerificationStatusPending, now, userID)

	return err
}

func (r *userRepository) FindPendingVerifications(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, telegram_id, phone_enc, snils_enc, dob_enc, full_name, 
		       username, role, verification_status, pd_consent_accepted, 
		       rejection_reason, created_at, updated_at
		FROM users 
		WHERE verification_status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, models.VerificationStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.TelegramID, &user.PhoneEncrypted, &user.SNILSEncrypted,
			&user.DOBEncrypted, &user.FullName, &user.Username, &user.Role,
			&user.VerificationStatus, &user.PDConsentAccepted, &user.RejectionReason,
			&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT id, telegram_id, phone_enc, snils_enc, dob_enc, full_name, 
		       username, role, verification_status, pd_consent_accepted, 
		       rejection_reason, created_at, updated_at
		FROM users WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.TelegramID, &user.PhoneEncrypted, &user.SNILSEncrypted,
		&user.DOBEncrypted, &user.FullName, &user.Username, &user.Role,
		&user.VerificationStatus, &user.PDConsentAccepted, &user.RejectionReason,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %s not found", userID)
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetWithDecryptedData(ctx context.Context, userID string) (*models.UserProfileWithPersonalData, error) {
	query := `
		SELECT u.id, u.telegram_id, u.full_name, u.username, u.role, 
		       u.verification_status, s.decrypted_snils, s.decrypted_dob, 
		       s.decrypted_phone, u.rejection_reason, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN (
			-- Subquery to decrypt sensitive data (this would be handled by the service layer)
			-- For now, we'll just select placeholders since decryption happens in the service
			SELECT $1 as user_id
		) s ON u.id = s.user_id
		WHERE u.id = $2
	`

	// This is a placeholder implementation. In the real implementation, 
	// the service layer would handle decryption before calling this method
	// or the database would have a special view/procedure for security officers
	var profile models.UserProfileWithPersonalData
	err := r.db.QueryRowContext(ctx, query, userID, userID).Scan(
		&profile.ID, &profile.TelegramID, &profile.FullName, &profile.Username,
		&profile.Role, &profile.VerificationStatus, &profile.SNILS, &profile.DateOfBirth,
		&profile.Phone, &profile.RejectionReason, &profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %s not found", userID)
		}
		return nil, err
	}

	return &profile, nil
}

func (r *userRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, telegram_id, phone_enc, snils_enc, dob_enc, full_name, 
		       username, role, verification_status, pd_consent_accepted, 
		       rejection_reason, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.TelegramID, &user.PhoneEncrypted, &user.SNILSEncrypted,
			&user.DOBEncrypted, &user.FullName, &user.Username, &user.Role,
			&user.VerificationStatus, &user.PDConsentAccepted, &user.RejectionReason,
			&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (r *userRepository) RevokeAccess(ctx context.Context, userID string) error {
	query := `
		UPDATE users 
		SET verification_status = $1, rejection_reason = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()
	reason := "Access revoked by administrator"
	_, err := r.db.ExecContext(ctx, query,
		models.VerificationStatusRejected, &reason, now, userID)

	return err
}