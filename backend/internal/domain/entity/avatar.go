package entity

import (
	"time"

	"github.com/google/uuid"
)

// Avatar represents the avatar domain entity
type Avatar struct {
	ID        string
	UserID    string
	PublicID  string // Cloudinary public_id
	PublicURL string // Cloudinary public URL
	SecureURL string // Cloudinary secure URL (HTTPS)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewAvatar creates a new avatar entity
func NewAvatar(userID, publicID, publicURL, secureURL string) *Avatar {
	return &Avatar{
		ID:        uuid.New().String(),
		UserID:    userID,
		PublicID:  publicID,
		PublicURL: publicURL,
		SecureURL: secureURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
