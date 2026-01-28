package utils

import (
	"errors"
	"net/http"

	domainErrors "backend/internal/domain/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
}

// ErrorData represents error details
type ErrorData struct {
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	response := Response{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = &ErrorData{
			Details: err.Error(),
		}
	}

	c.JSON(statusCode, response)
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c *gin.Context, err error) {
	var validationErrors []string
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			validationErrors = append(validationErrors, formatValidationError(fe))
		}
	}

	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: "validation failed",
		Error: &ErrorData{
			Code:    "VALIDATION_ERROR",
			Details: validationErrors,
		},
	})
}

// HandleDomainError handles domain-specific errors
func HandleDomainError(c *gin.Context, err error) {
	var domainErr *domainErrors.DomainError
	if errors.As(err, &domainErr) {
		statusCode := getStatusCodeFromDomainError(domainErr)
		c.JSON(statusCode, Response{
			Success: false,
			Message: domainErr.Message,
			Error: &ErrorData{
				Code: domainErr.Code,
			},
		})
		return
	}

	// Unknown error
	ErrorResponse(c, http.StatusInternalServerError, "internal server error", err)
}

// getStatusCodeFromDomainError maps domain errors to HTTP status codes
func getStatusCodeFromDomainError(err *domainErrors.DomainError) int {
	switch err.Code {
	case "USER_NOT_FOUND":
		return http.StatusNotFound
	case "USER_EXISTS":
		return http.StatusConflict
	case "INVALID_CREDENTIALS":
		return http.StatusUnauthorized
	case "UNAUTHORIZED", "INVALID_TOKEN":
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// formatValidationError formats a validation error
func formatValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	default:
		return fe.Field() + " is invalid"
	}
}
