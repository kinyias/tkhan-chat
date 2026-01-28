package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"backend/internal/domain/entity"
	"backend/internal/domain/repository"
)

// OAuthUseCase defines the interface for OAuth use cases
type OAuthUseCase interface {
	GenerateStateToken() (string, error)
	GetGoogleAuthURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*entity.User, error)
}

type oauthUseCase struct {
	userRepo     repository.UserRepository
	oauthService OAuthService
}

// NewOAuthUseCase creates a new OAuth use case
func NewOAuthUseCase(userRepo repository.UserRepository, oauthService OAuthService) OAuthUseCase {
	return &oauthUseCase{
		userRepo:     userRepo,
		oauthService: oauthService,
	}
}

// GenerateStateToken generates a random state token for CSRF protection
func (uc *oauthUseCase) GenerateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetGoogleAuthURL returns the Google OAuth authorization URL
func (uc *oauthUseCase) GetGoogleAuthURL(state string) string {
	return uc.oauthService.GetAuthURL(state)
}

// HandleGoogleCallback handles the Google OAuth callback
func (uc *oauthUseCase) HandleGoogleCallback(ctx context.Context, code string) (*entity.User, error) {
	// Exchange code for token
	token, err := uc.oauthService.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	userInfo, err := uc.oauthService.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check if user already exists by OAuth ID
	existingUser, err := uc.userRepo.GetByOAuthID(ctx, "google", userInfo.ID)
	if err == nil {
		// User exists, return it
		return existingUser, nil
	}

	// Check if user exists by email (linking existing account)
	existingUser, err = uc.userRepo.GetByEmail(ctx, userInfo.Email)
	if err == nil {
		// User exists with this email, link OAuth account
		existingUser.OAuthProvider = "google"
		existingUser.OAuthID = userInfo.ID
		if existingUser.Avatar == "" {
			existingUser.Avatar = userInfo.Picture
		}
		if err := uc.userRepo.Update(ctx, existingUser); err != nil {
			return nil, fmt.Errorf("failed to link OAuth account: %w", err)
		}
		return existingUser, nil
	}

	// Create new user
	newUser := entity.NewOAuthUser(
		userInfo.Email,
		userInfo.Name,
		userInfo.Picture,
		"google",
		userInfo.ID,
	)

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}
