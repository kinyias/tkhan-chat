package errors

import "fmt"

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Common domain errors
var (
	ErrUserNotFound       = &DomainError{Code: "USER_NOT_FOUND", Message: "user not found"}
	ErrUserExists         = &DomainError{Code: "USER_EXISTS", Message: "user with this email already exists"}
	ErrInvalidCredentials = &DomainError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"}
	ErrUnauthorized       = &DomainError{Code: "UNAUTHORIZED", Message: "unauthorized access"}
	ErrInvalidToken       = &DomainError{Code: "INVALID_TOKEN", Message: "invalid or expired token"}
	ErrTokenRevoked       = &DomainError{Code: "TOKEN_REVOKED", Message: "token has been revoked"}
	ErrTokenExpired       = &DomainError{Code: "TOKEN_EXPIRED", Message: "token has expired"}
	ErrRefreshTokenNotFound = &DomainError{Code: "REFRESH_TOKEN_NOT_FOUND", Message: "refresh token not found"}
)
