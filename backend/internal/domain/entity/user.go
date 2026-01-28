package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user domain entity
type User struct {
	ID        string
	Email     string
	Password  string // bcrypt hashed
	Name      string
	Avatar    string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a new user entity
func NewUser(email, password, name, phone string) *User {
	return &User{
		ID:        uuid.New().String(),
		Email:     email,
		Password:  password,
		Name:      name,
		Avatar:    "",
		Phone:     phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
