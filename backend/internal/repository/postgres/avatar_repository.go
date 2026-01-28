package postgres

import (
	"context"

	"backend/internal/domain/entity"
	"backend/internal/domain/errors"
	"backend/internal/domain/repository"

	"gorm.io/gorm"
)

// AvatarModel represents the GORM database model for avatars
type AvatarModel struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	UserID    string `gorm:"uniqueIndex;not null;type:uuid"`
	PublicID  string `gorm:"not null"`
	PublicURL string `gorm:"type:text;not null"`
	SecureURL string `gorm:"type:text;not null"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
}

// TableName specifies the table name for AvatarModel
func (AvatarModel) TableName() string {
	return "avatars"
}

type avatarRepository struct {
	db *gorm.DB
}

// NewAvatarRepository creates a new avatar repository
func NewAvatarRepository(db *gorm.DB) repository.AvatarRepository {
	return &avatarRepository{db: db}
}

func (r *avatarRepository) Create(ctx context.Context, avatar *entity.Avatar) error {
	model := r.toModel(avatar)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *avatarRepository) GetByUserID(ctx context.Context, userID string) (*entity.Avatar, error) {
	var model AvatarModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(&model), nil
}

func (r *avatarRepository) Update(ctx context.Context, avatar *entity.Avatar) error {
	model := r.toModel(avatar)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *avatarRepository) Delete(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Delete(&AvatarModel{}, "user_id = ?", userID).Error
}

// toModel converts domain entity to GORM model
func (r *avatarRepository) toModel(avatar *entity.Avatar) *AvatarModel {
	return &AvatarModel{
		ID:        avatar.ID,
		UserID:    avatar.UserID,
		PublicID:  avatar.PublicID,
		PublicURL: avatar.PublicURL,
		SecureURL: avatar.SecureURL,
	}
}

// toEntity converts GORM model to domain entity
func (r *avatarRepository) toEntity(model *AvatarModel) *entity.Avatar {
	return &entity.Avatar{
		ID:        model.ID,
		UserID:    model.UserID,
		PublicID:  model.PublicID,
		PublicURL: model.PublicURL,
		SecureURL: model.SecureURL,
	}
}
