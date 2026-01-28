package repository

import (
	"context"

	"backend/internal/domain/entity"
)

// AvatarRepository defines the interface for avatar data access
type AvatarRepository interface {
	Create(ctx context.Context, avatar *entity.Avatar) error
	GetByUserID(ctx context.Context, userID string) (*entity.Avatar, error)
	Update(ctx context.Context, avatar *entity.Avatar) error
	Delete(ctx context.Context, userID string) error
}
