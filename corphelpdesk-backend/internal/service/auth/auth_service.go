package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yourcompany/corphelpdesk-backend/internal/models"
	"github.com/yourcompany/corphelpdesk-backend/internal/repository/postgres"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/encryption"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/validation"
)

type AuthService interface {
	VerifyTelegramAuth(ctx context.Context, telegramID int64) (*models.User, error)
	GetProfile(ctx context.Context, userID string) (*models.User, error)
	UpdateProfile(ctx context.Context, userID string, profile *models.UserProfileUpdate) error
	GetProfileWithPersonalData(ctx context.Context, userID string) (*models.UserProfileWithPersonalData, error)
}

type authService struct {
	userRepo    postgres.UserRepository
	encService  *encryption.AESEncryption
}

func NewAuthService(userRepo postgres.UserRepository, encService *encryption.AESEncryption) AuthService {
	return &authService{
		userRepo:   userRepo,
		encService: encService,
	}
}

func (s *authService) VerifyTelegramAuth(ctx context.Context, telegramID int64) (*models.User, error) {
	// Try to find existing user by telegram ID
	user, err := s.userRepo.FindByTelegramID(ctx, telegramID)
	if err != nil {
		// If user doesn't exist, create a new one with pending status
		if errors.Is(err, fmt.Errorf("user with telegram_id %d not found", telegramID)) {
			newUser := &models.User{
				TelegramID:         telegramID,
				FullName:          "", // Will be populated later
				Role:              models.UserRoleClient,
				VerificationStatus: models.VerificationStatusPending,
				PDConsentAccepted: false,
			}

			if err := s.userRepo.Create(ctx, newUser); err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}

			return newUser, nil
		}
		return nil, err
	}

	return user, nil
}

func (s *authService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *authService) UpdateProfile(ctx context.Context, userID string, profile *models.UserProfileUpdate) error {
	// Validate SNILS
	if err := validation.ValidateSNILS(profile.SNILS); err != nil {
		return fmt.Errorf("invalid SNILS: %w", err)
	}

	// Encrypt sensitive data
	snilsEncrypted, err := s.encService.Encrypt(profile.SNILS)
	if err != nil {
		return fmt.Errorf("failed to encrypt SNILS: %w", err)
	}

	phoneEncrypted, err := s.encService.Encrypt(profile.Phone)
	if err != nil {
		return fmt.Errorf("failed to encrypt phone: %w", err)
	}

	dobEncrypted, err := s.encService.Encrypt(profile.DateOfBirth.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("failed to encrypt date of birth: %w", err)
	}

	// Update user profile
	err = s.userRepo.UpdateUserProfile(ctx, userID, snilsEncrypted, phoneEncrypted, dobEncrypted, profile.PDConsent)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (s *authService) GetProfileWithPersonalData(ctx context.Context, userID string) (*models.UserProfileWithPersonalData, error) {
	// First, get the regular user data
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Then get the user data with decrypted personal information
	profile, err := s.userRepo.GetWithDecryptedData(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive data
	if user.SNILSEncrypted != nil {
		snils, err := s.encService.Decrypt(user.SNILSEncrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt SNILS: %w", err)
		}
		profile.SNILS = snils
	}

	if user.DOBEncrypted != nil {
		dob, err := s.encService.Decrypt(user.DOBEncrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt date of birth: %w", err)
		}
		profile.DateOfBirth = dob
	}

	if user.PhoneEncrypted != nil {
		phone, err := s.encService.Decrypt(user.PhoneEncrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt phone: %w", err)
		}
		profile.Phone = phone
	}

	return profile, nil
}