package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"backend/internal/domain/entity"
	"backend/internal/domain/errors"
	"backend/internal/domain/repository"
	"backend/internal/infrastructure/email"

	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase defines the interface for authentication use cases
type AuthUseCase interface {
	Register(ctx context.Context, email, password, name, phone string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, email string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type authUseCase struct {
	userRepo     repository.UserRepository
	emailService email.EmailService
}

// NewAuthUseCase creates a new authentication use case
func NewAuthUseCase(userRepo repository.UserRepository, emailService email.EmailService) AuthUseCase {
	return &authUseCase{
		userRepo:     userRepo,
		emailService: emailService,
	}
}

// Register creates a new user account
func (uc *authUseCase) Register(ctx context.Context, email, password, name, phone string) (*entity.User, error) {
	// Check if user already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := entity.NewUser(email, string(hashedPassword), name, phone)

	// Generate verification token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	user.VerificationToken = token
	user.VerificationTokenExpiresAt = time.Now().Add(24 * time.Hour) // 24 hours

	// Save user to database
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Send verification email
	if err := uc.emailService.SendVerificationEmail(user.Email, user.Name, token); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	return user, nil
}

// Login authenticates a user
func (uc *authUseCase) Login(ctx context.Context, email, password string) (*entity.User, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Check if user is OAuth user (no password)
	if user.IsOAuthUser() {
		return nil, fmt.Errorf("this account uses OAuth login, please use Google login")
	}

	// Check if email is verified
	if !user.EmailVerified {
		return nil, errors.ErrEmailNotVerified
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	return user, nil
}

// VerifyEmail verifies a user's email address
func (uc *authUseCase) VerifyEmail(ctx context.Context, token string) error {
	// Find user by verification token
	user, err := uc.userRepo.GetByVerificationToken(ctx, token)
	if err != nil {
		return errors.ErrInvalidVerificationToken
	}

	// Check if token is expired
	if time.Now().After(user.VerificationTokenExpiresAt) {
		return errors.ErrVerificationTokenExpired
	}

	// Check if already verified
	if user.EmailVerified {
		return nil // Already verified, no error
	}

	// Mark email as verified and clear token
	user.EmailVerified = true
	user.VerificationToken = ""
	user.VerificationTokenExpiresAt = time.Time{}

	// Update user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ResendVerificationEmail resends the verification email
func (uc *authUseCase) ResendVerificationEmail(ctx context.Context, email string) error {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Check if already verified
	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	// Generate new verification token
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	user.VerificationToken = token
	user.VerificationTokenExpiresAt = time.Now().Add(24 * time.Hour)

	// Update user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Send verification email
	if err := uc.emailService.SendVerificationEmail(user.Email, user.Name, token); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// ForgotPassword initiates the password reset process
func (uc *authUseCase) ForgotPassword(ctx context.Context, email string) error {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// Check if user is OAuth user
	if user.IsOAuthUser() {
		// Don't send reset email for OAuth users
		return nil
	}

	// Generate reset token
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	user.ResetPasswordToken = token
	user.ResetPasswordTokenExpiresAt = time.Now().Add(1 * time.Hour) // 1 hour

	// Update user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Send password reset email
	if err := uc.emailService.SendPasswordResetEmail(user.Email, user.Name, token); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// ResetPassword resets a user's password
func (uc *authUseCase) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Find user by reset token
	user, err := uc.userRepo.GetByResetPasswordToken(ctx, token)
	if err != nil {
		return errors.ErrInvalidResetToken
	}

	// Check if token is expired
	if time.Now().After(user.ResetPasswordTokenExpiresAt) {
		return errors.ErrResetTokenExpired
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and clear reset token
	user.Password = string(hashedPassword)
	user.ResetPasswordToken = ""
	user.ResetPasswordTokenExpiresAt = time.Time{}

	// Update user
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// generateToken generates a random token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
