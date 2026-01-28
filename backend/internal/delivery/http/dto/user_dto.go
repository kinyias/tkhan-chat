package dto

import "time"

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
}

// LoginRequest represents the user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserRequest represents the user update request
type UpdateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Phone string `json:"phone"`
}

// AvatarDTO represents avatar data transfer object
type AvatarDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PublicID  string    `json:"public_id"`
	PublicURL string    `json:"public_url"`
	SecureURL string    `json:"secure_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserResponse represents the user response
type UserResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Avatar    *AvatarDTO `json:"avatar,omitempty"`
	Phone     string     `json:"phone"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// LoginResponse represents the login response with tokens
type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	User         *UserResponse `json:"user"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ListUsersResponse represents the list users response
type ListUsersResponse struct {
	Users  []*UserResponse `json:"users"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// OAuthCallbackRequest represents the OAuth callback request
type OAuthCallbackRequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
}

// OAuthAuthURLResponse represents the OAuth authorization URL response
type OAuthAuthURLResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

