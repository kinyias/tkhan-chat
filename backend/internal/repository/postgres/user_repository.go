package postgres

import (
	"context"
	"time"

	"backend/internal/domain/entity"
	"backend/internal/domain/errors"
	"backend/internal/domain/repository"

	"gorm.io/gorm"
)

// UserModel represents the GORM database model for users
type UserModel struct {
	ID            string `gorm:"primaryKey;type:uuid"`
	Email         string `gorm:"uniqueIndex;not null"`
	Password      string
	Name          string `gorm:"not null"`
	Avatar        string
	Phone         string
	OAuthProvider string `gorm:"column:oauth_provider"`
	OAuthID       string `gorm:"column:oauth_id"`
	CreatedAt     int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt     int64  `gorm:"autoUpdateTime:milli"`
}

// TableName specifies the table name for UserModel
func (UserModel) TableName() string {
	return "users"
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	model := r.toModel(user)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(&model), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(&model), nil
}

func (r *userRepository) GetByOAuthID(ctx context.Context, provider, oauthID string) (*entity.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("oauth_provider = ? AND oauth_id = ?", provider, oauthID).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(&model), nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	model := r.toModel(user)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id).Error
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	var models []UserModel
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]*entity.User, len(models))
	for i, model := range models {
		users[i] = r.toEntity(&model)
	}
	return users, nil
}

// toModel converts domain entity to GORM model
func (r *userRepository) toModel(user *entity.User) *UserModel {
	return &UserModel{
		ID:            user.ID,
		Email:         user.Email,
		Password:      user.Password,
		Name:          user.Name,
		Avatar:        user.Avatar,
		Phone:         user.Phone,
		OAuthProvider: user.OAuthProvider,
		OAuthID:       user.OAuthID,
	}
}

// toEntity converts GORM model to domain entity
func (r *userRepository) toEntity(model *UserModel) *entity.User {
	return &entity.User{
		ID:            model.ID,
		Email:         model.Email,
		Password:      model.Password,
		Name:          model.Name,
		Avatar:        model.Avatar,
		Phone:         model.Phone,
		OAuthProvider: model.OAuthProvider,
		OAuthID:       model.OAuthID,
		CreatedAt:     time.UnixMilli(model.CreatedAt),
		UpdatedAt:     time.UnixMilli(model.UpdatedAt),
	}
}
