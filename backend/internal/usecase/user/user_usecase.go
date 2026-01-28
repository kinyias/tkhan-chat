package user

import (
	"context"
	"time"

	"backend/internal/domain/entity"
	"backend/internal/domain/errors"
	"backend/internal/domain/repository"

	"golang.org/x/crypto/bcrypt"
)

// UserUseCase defines the interface for user business logic
type UserUseCase interface {
	Register(ctx context.Context, email, password, name, phone string) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Authenticate(ctx context.Context, email, password string) (*entity.User, error)
	Update(ctx context.Context, id, name, avatar, phone string) (*entity.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, error)
}

type userUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
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

func (uc *userUseCase) Update(ctx context.Context, id, name, avatar, phone string) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Name = name
	user.Avatar = avatar
	user.Phone = phone
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUseCase) Delete(ctx context.Context, id string) error {
	return uc.userRepo.Delete(ctx, id)
}

func (uc *userUseCase) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	return uc.userRepo.List(ctx, limit, offset)
}
