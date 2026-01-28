package cloudinary

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Service defines the interface for Cloudinary operations
type Service interface {
	UploadAvatar(ctx context.Context, file multipart.File, userID string) (*UploadResult, error)
	DeleteAvatar(ctx context.Context, publicID string) error
}

// UploadResult contains the result of a Cloudinary upload
type UploadResult struct {
	PublicID  string
	PublicURL string
	SecureURL string
}

type service struct {
	cld *cloudinary.Cloudinary
}

// NewService creates a new Cloudinary service
func NewService(cloudName, apiKey, apiSecret string) (Service, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	return &service{
		cld: cld,
	}, nil
}

// UploadAvatar uploads an avatar image to Cloudinary
func (s *service) UploadAvatar(ctx context.Context, file multipart.File, userID string) (*UploadResult, error) {
	overwrite := true
	// Upload the file to Cloudinary
	uploadParams := uploader.UploadParams{
		Folder:         "avatars",
		PublicID:       fmt.Sprintf("user_%s", userID),
		Overwrite:      &overwrite,
		ResourceType:   "image",
		Transformation: "c_fill,g_face,h_400,w_400", // Crop to 400x400 focusing on face
	}

	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	return &UploadResult{
		PublicID:  result.PublicID,
		PublicURL: result.URL,
		SecureURL: result.SecureURL,
	}, nil
}

// DeleteAvatar deletes an avatar from Cloudinary
func (s *service) DeleteAvatar(ctx context.Context, publicID string) error {
	if publicID == "" {
		return nil // Nothing to delete
	}

	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})

	if err != nil {
		return fmt.Errorf("failed to delete avatar: %w", err)
	}

	return nil
}
