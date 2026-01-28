package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user domain entity
type User struct {
	ID            string
	Email         string
	Password      string // bcrypt hashed (optional for OAuth users)
	Name          string
	Avatar        string
	Phone         string
	OAuthProvider string // e.g., "google", "facebook", etc.
	OAuthID       string // OAuth provider's user ID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewUser creates a new user entity
func NewUser(email, password, name, phone string) *User {
	return &User{
		ID:            uuid.New().String(),
		Email:         email,
		Password:      password,
		Name:          name,
		Avatar:        "",
		Phone:         phone,
		OAuthProvider: "",
		OAuthID:       "",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// NewOAuthUser creates a new OAuth user entity
func NewOAuthUser(email, name, avatar, provider, oauthID string) *User {
	return &User{
		ID:            uuid.New().String(),
		Email:         email,
		Password:      "", // No password for OAuth users
		Name:          name,
		Avatar:        avatar,
		Phone:         "",
		OAuthProvider: provider,
		OAuthID:       oauthID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// IsOAuthUser checks if the user is an OAuth user
func (u *User) IsOAuthUser() bool {
	return u.OAuthProvider != "" && u.OAuthID != ""
}

