package user

import (
	"context"
	"mime/multipart"
	"time"

	"backend/internal/domain/entity"
	"backend/internal/domain/errors"
	"backend/internal/domain/repository"
	"backend/internal/infrastructure/cloudinary"

	"golang.org/x/crypto/bcrypt"
)

// UserUseCase defines the interface for user business logic
type UserUseCase interface {
	Register(ctx context.Context, email, password, name, phone string) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Authenticate(ctx context.Context, email, password string) (*entity.User, error)
	Update(ctx context.Context, id, name, phone string) (*entity.User, error)
	UpdateAvatar(ctx context.Context, userID string, file multipart.File) (*entity.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, error)
}

type userUseCase struct {
	userRepo       repository.UserRepository
	avatarRepo     repository.AvatarRepository
	cloudinaryServ cloudinary.Service
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(
	userRepo repository.UserRepository,
	avatarRepo repository.AvatarRepository,
	cloudinaryServ cloudinary.Service,
) UserUseCase {
	return &userUseCase{
		userRepo:       userRepo,
		avatarRepo:     avatarRepo,
		cloudinaryServ: cloudinaryServ,
	}
}

func (uc *userUseCase) Register(ctx context.Context, email, password, name, phone string) (*entity.User, error) {
	// Check if user already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, &errors.DomainError{
			Code:    "PASSWORD_HASH_FAILED",
			Message: "failed to hash password",
			Err:     err,
		}
	}

	// Create user entity
	user := entity.NewUser(email, string(hashedPassword), name, phone)

	// Save to repository
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUseCase) GetByID(ctx context.Context, id string) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *userUseCase) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return uc.userRepo.GetByEmail(ctx, email)
}

func (uc *userUseCase) Authenticate(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	return user, nil
}

func (uc *userUseCase) Update(ctx context.Context, id, name, phone string) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Name = name
	user.Phone = phone
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateAvatar handles avatar upload with automatic deletion of old avatar
func (uc *userUseCase) UpdateAvatar(ctx context.Context, userID string, file multipart.File) (*entity.User, error) {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get existing avatar (if any)
	existingAvatar, _ := uc.avatarRepo.GetByUserID(ctx, userID)

	// Upload new avatar to Cloudinary
	uploadResult, err := uc.cloudinaryServ.UploadAvatar(ctx, file, userID)
	if err != nil {
		return nil, &errors.DomainError{
			Code:    "AVATAR_UPLOAD_FAILED",
			Message: "failed to upload avatar",
			Err:     err,
		}
	}

	// Create new avatar entity
	newAvatar := entity.NewAvatar(userID, uploadResult.PublicID, uploadResult.PublicURL, uploadResult.SecureURL)

	// Save or update avatar in database
	if existingAvatar != nil {
		// Update existing avatar
		newAvatar.ID = existingAvatar.ID
		if err := uc.avatarRepo.Update(ctx, newAvatar); err != nil {
			return nil, err
		}

		// Delete old avatar from Cloudinary (if it has a public_id)
		if existingAvatar.PublicID != "" {
			// Delete in background, don't fail if deletion fails
			go func() {
				_ = uc.cloudinaryServ.DeleteAvatar(context.Background(), existingAvatar.PublicID)
			}()
		}
	} else {
		// Create new avatar
		if err := uc.avatarRepo.Create(ctx, newAvatar); err != nil {
			return nil, err
		}
	}

	// Update user's avatar reference
	user.Avatar = newAvatar
	user.UpdatedAt = time.Now()

	return user, nil
}

func (uc *userUseCase) Delete(ctx context.Context, id string) error {
	// Get user to check if they have an avatar
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete avatar from Cloudinary if exists
	if user.Avatar != nil && user.Avatar.PublicID != "" {
		// Delete in background, don't fail if deletion fails
		go func() {
			_ = uc.cloudinaryServ.DeleteAvatar(context.Background(), user.Avatar.PublicID)
		}()
	}

	// Delete avatar from database (cascade will handle this via foreign key)
	// Delete user from database
	return uc.userRepo.Delete(ctx, id)
}

func (uc *userUseCase) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	return uc.userRepo.List(ctx, limit, offset)
}
