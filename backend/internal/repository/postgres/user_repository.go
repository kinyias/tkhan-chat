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
	ID                           string `gorm:"primaryKey;type:uuid"`
	Email                        string `gorm:"uniqueIndex;not null"`
	Password                     string
	Name                         string `gorm:"not null"`
	Phone                        string
	OAuthProvider                string `gorm:"column:oauth_provider"`
	OAuthID                      string `gorm:"column:oauth_id"`
	EmailVerified                bool   `gorm:"default:false"`
	VerificationToken            string `gorm:"column:verification_token"`
	VerificationTokenExpiresAt   int64  `gorm:"column:verification_token_expires_at"`
	ResetPasswordToken           string `gorm:"column:reset_password_token"`
	ResetPasswordTokenExpiresAt  int64  `gorm:"column:reset_password_token_expires_at"`
	CreatedAt                    int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt                    int64  `gorm:"autoUpdateTime:milli"`
}

// TableName specifies the table name for UserModel
func (UserModel) TableName() string {
	return "users"
}

type userRepository struct {
	db         *gorm.DB
	avatarRepo repository.AvatarRepository
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, avatarRepo repository.AvatarRepository) repository.UserRepository {
	return &userRepository{
		db:         db,
		avatarRepo: avatarRepo,
	}
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
	return r.toEntity(ctx, &model), nil
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
	return r.toEntity(ctx, &model), nil
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
	return r.toEntity(ctx, &model), nil
}

func (r *userRepository) GetByVerificationToken(ctx context.Context, token string) (*entity.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("verification_token = ?", token).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(ctx, &model), nil
}

func (r *userRepository) GetByResetPasswordToken(ctx context.Context, token string) (*entity.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("reset_password_token = ?", token).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(ctx, &model), nil
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
		users[i] = r.toEntity(ctx, &model)
	}
	return users, nil
}

// toModel converts domain entity to GORM model
func (r *userRepository) toModel(user *entity.User) *UserModel {
	var verificationTokenExpiresAt, resetPasswordTokenExpiresAt int64
	if !user.VerificationTokenExpiresAt.IsZero() {
		verificationTokenExpiresAt = user.VerificationTokenExpiresAt.UnixMilli()
	}
	if !user.ResetPasswordTokenExpiresAt.IsZero() {
		resetPasswordTokenExpiresAt = user.ResetPasswordTokenExpiresAt.UnixMilli()
	}

	return &UserModel{
		ID:                           user.ID,
		Email:                        user.Email,
		Password:                     user.Password,
		Name:                         user.Name,
		Phone:                        user.Phone,
		OAuthProvider:                user.OAuthProvider,
		OAuthID:                      user.OAuthID,
		EmailVerified:                user.EmailVerified,
		VerificationToken:            user.VerificationToken,
		VerificationTokenExpiresAt:   verificationTokenExpiresAt,
		ResetPasswordToken:           user.ResetPasswordToken,
		ResetPasswordTokenExpiresAt:  resetPasswordTokenExpiresAt,
	}
}

// toEntity converts GORM model to domain entity
func (r *userRepository) toEntity(ctx context.Context, model *UserModel) *entity.User {
	// Try to load avatar from avatar repository
	var avatar *entity.Avatar
	if r.avatarRepo != nil {
		avatar, _ = r.avatarRepo.GetByUserID(ctx, model.ID)
		// Ignore error if avatar not found, it's optional
	}

	var verificationTokenExpiresAt, resetPasswordTokenExpiresAt time.Time
	if model.VerificationTokenExpiresAt > 0 {
		verificationTokenExpiresAt = time.UnixMilli(model.VerificationTokenExpiresAt)
	}
	if model.ResetPasswordTokenExpiresAt > 0 {
		resetPasswordTokenExpiresAt = time.UnixMilli(model.ResetPasswordTokenExpiresAt)
	}

	return &entity.User{
		ID:                           model.ID,
		Email:                        model.Email,
		Password:                     model.Password,
		Name:                         model.Name,
		Avatar:                       avatar,
		Phone:                        model.Phone,
		OAuthProvider:                model.OAuthProvider,
		OAuthID:                      model.OAuthID,
		EmailVerified:                model.EmailVerified,
		VerificationToken:            model.VerificationToken,
		VerificationTokenExpiresAt:   verificationTokenExpiresAt,
		ResetPasswordToken:           model.ResetPasswordToken,
		ResetPasswordTokenExpiresAt:  resetPasswordTokenExpiresAt,
		CreatedAt:                    time.UnixMilli(model.CreatedAt),
		UpdatedAt:                    time.UnixMilli(model.UpdatedAt),
	}
}
