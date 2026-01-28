package auth

import (
	"context"
	"time"

	"backend/internal/domain/entity"
	"backend/internal/domain/errors"
	"backend/internal/domain/repository"
)

// RefreshTokenUseCase defines the interface for refresh token operations
type RefreshTokenUseCase interface {
	CreateRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error
	ValidateRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
}

type refreshTokenUseCase struct {
	refreshTokenRepo repository.RefreshTokenRepository
}

// NewRefreshTokenUseCase creates a new refresh token use case
func NewRefreshTokenUseCase(refreshTokenRepo repository.RefreshTokenRepository) RefreshTokenUseCase {
	return &refreshTokenUseCase{
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (uc *refreshTokenUseCase) CreateRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error {
	refreshToken := entity.NewRefreshToken(userID, token, expiresAt)
	return uc.refreshTokenRepo.Create(ctx, refreshToken)
}

func (uc *refreshTokenUseCase) ValidateRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	refreshToken, err := uc.refreshTokenRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.ErrRefreshTokenNotFound
	}

	if !refreshToken.IsValid() {
		if refreshToken.RevokedAt != nil {
			return nil, errors.ErrTokenRevoked
		}
		return nil, errors.ErrTokenExpired
	}

	return refreshToken, nil
}

func (uc *refreshTokenUseCase) RevokeRefreshToken(ctx context.Context, token string) error {
	return uc.refreshTokenRepo.Revoke(ctx, token)
}

func (uc *refreshTokenUseCase) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return uc.refreshTokenRepo.RevokeAllByUserID(ctx, userID)
}
