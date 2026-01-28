package auth

import (
	"time"

	"backend/internal/domain/errors"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID    string    `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTService defines the interface for JWT operations
type JWTService interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateToken(tokenString string, expectedType TokenType) (*JWTClaims, error)
	GetAccessTokenExpiration() time.Duration
	GetRefreshTokenExpiration() time.Duration
}

type jwtService struct {
	secretKey                string
	accessTokenExpireMinutes int
	refreshTokenExpireDays   int
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessTokenExpireMinutes, refreshTokenExpireDays int) JWTService {
	return &jwtService{
		secretKey:                secretKey,
		accessTokenExpireMinutes: accessTokenExpireMinutes,
		refreshTokenExpireDays:   refreshTokenExpireDays,
	}
}

func (s *jwtService) GenerateAccessToken(userID string) (string, error) {
	return s.generateToken(userID, AccessToken, time.Minute*time.Duration(s.accessTokenExpireMinutes))
}

func (s *jwtService) GenerateRefreshToken(userID string) (string, error) {
	return s.generateToken(userID, RefreshToken, time.Hour*24*time.Duration(s.refreshTokenExpireDays))
}

func (s *jwtService) generateToken(userID string, tokenType TokenType, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *jwtService) ValidateToken(tokenString string, expectedType TokenType) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrInvalidToken
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	// Verify token type
	if claims.TokenType != expectedType {
		return nil, errors.ErrInvalidToken
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.ErrTokenExpired
	}

	return claims, nil
}

func (s *jwtService) GetAccessTokenExpiration() time.Duration {
	return time.Minute * time.Duration(s.accessTokenExpireMinutes)
}

func (s *jwtService) GetRefreshTokenExpiration() time.Duration {
	return time.Hour * 24 * time.Duration(s.refreshTokenExpireDays)
}
