package user_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/domain/entity"
	"backend/internal/domain/errors"
	"backend/internal/usecase/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func TestRegister_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := user.NewUserUseCase(mockRepo)

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.ErrUserNotFound)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

	result, err := uc.Register(context.Background(), "test@example.com", "password123", "Test User", "1234567890")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "Test User", result.Name)
	assert.NotEmpty(t, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestRegister_UserExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := user.NewUserUseCase(mockRepo)

	existingUser := &entity.User{
		ID:        "123",
		Email:     "test@example.com",
		Name:      "Existing User",
		CreatedAt: time.Now(),
	}
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	result, err := uc.Register(context.Background(), "test@example.com", "password123", "Test User", "1234567890")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errors.ErrUserExists, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := user.NewUserUseCase(mockRepo)

	// Pre-hashed password for "password123"
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	existingUser := &entity.User{
		ID:       "123",
		Email:    "test@example.com",
		Password: hashedPassword,
		Name:     "Test User",
	}

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	result, err := uc.Authenticate(context.Background(), "test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test@example.com", result.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := user.NewUserUseCase(mockRepo)

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.ErrUserNotFound)

	result, err := uc.Authenticate(context.Background(), "test@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errors.ErrInvalidCredentials, err)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := user.NewUserUseCase(mockRepo)

	expectedUser := &entity.User{
		ID:    "123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	mockRepo.On("GetByID", mock.Anything, "123").Return(expectedUser, nil)

	result, err := uc.GetByID(context.Background(), "123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Email, result.Email)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := user.NewUserUseCase(mockRepo)

	mockRepo.On("GetByID", mock.Anything, "999").Return(nil, errors.ErrUserNotFound)

	result, err := uc.GetByID(context.Background(), "999")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errors.ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}
