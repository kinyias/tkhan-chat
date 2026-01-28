package postgres

import (
	"context"
	"time"

	"backend/internal/domain/entity"
	"backend/internal/domain/errors"
	"backend/internal/domain/repository"

	"gorm.io/gorm"
)

// RefreshTokenModel represents the GORM database model for refresh tokens
type RefreshTokenModel struct {
	ID        string     `gorm:"primaryKey;type:uuid"`
	UserID    string     `gorm:"type:uuid;not null;index"`
	Token     string     `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	RevokedAt *time.Time `gorm:"default:null"`
}

// TableName specifies the table name for RefreshTokenModel
func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	model := r.toModel(token)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	var model RefreshTokenModel
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.ErrRefreshTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(&model), nil
}

func (r *refreshTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.RefreshToken, error) {
	var models []RefreshTokenModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	tokens := make([]*entity.RefreshToken, len(models))
	for i, model := range models {
		tokens[i] = r.toEntity(&model)
	}
	return tokens, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("token = ?", token).
		Update("revoked_at", now).Error
}

func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&RefreshTokenModel{}).Error
}

// toModel converts domain entity to GORM model
func (r *refreshTokenRepository) toModel(token *entity.RefreshToken) *RefreshTokenModel {
	model := &RefreshTokenModel{
		ID:        token.ID,
		UserID:    token.UserID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
	}
	
	if token.RevokedAt != nil {
		revokedAt := *token.RevokedAt
		model.RevokedAt = &revokedAt
	}
	
	return model
}

// toEntity converts GORM model to domain entity
func (r *refreshTokenRepository) toEntity(model *RefreshTokenModel) *entity.RefreshToken {
	token := &entity.RefreshToken{
		ID:        model.ID,
		UserID:    model.UserID,
		Token:     model.Token,
		ExpiresAt: model.ExpiresAt,
		CreatedAt: model.CreatedAt,
	}
	
	if model.RevokedAt != nil {
		revokedAt := *model.RevokedAt
		token.RevokedAt = &revokedAt
	}
	
	return token
}
