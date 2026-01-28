package handler

import (
	"net/http"
	"time"

	"backend/internal/delivery/http/dto"
	"backend/internal/domain/errors"
	"backend/internal/usecase/auth"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	authUseCase         auth.AuthUseCase
	jwtService          auth.JWTService
	refreshTokenUseCase auth.RefreshTokenUseCase
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authUseCase auth.AuthUseCase,
	jwtService auth.JWTService,
	refreshTokenUseCase auth.RefreshTokenUseCase,
) *AuthHandler {
	return &AuthHandler{
		authUseCase:         authUseCase,
		jwtService:          jwtService,
		refreshTokenUseCase: refreshTokenUseCase,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user account and send verification email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	user, err := h.authUseCase.Register(c.Request.Context(), req.Email, req.Password, req.Name, req.Phone)
	if err != nil {
		if err == errors.ErrUserAlreadyExists {
			utils.ErrorResponse(c, http.StatusConflict, "user already exists", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to register user", err)
		return
	}

	userResponse := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	utils.SuccessResponse(c, http.StatusCreated, "registration successful, please check your email to verify your account", userResponse)
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	user, err := h.authUseCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == errors.ErrInvalidCredentials {
			utils.ErrorResponse(c, http.StatusUnauthorized, "invalid credentials", err)
			return
		}
		if err == errors.ErrEmailNotVerified {
			utils.ErrorResponse(c, http.StatusForbidden, "email not verified", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to login", err)
		return
	}

	// Generate JWT tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate access token", err)
		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate refresh token", err)
		return
	}

	// Store refresh token
	expiresAt := time.Now().Add(h.jwtService.GetRefreshTokenExpiration())
	if err := h.refreshTokenUseCase.CreateRefreshToken(c.Request.Context(), user.ID, refreshToken, expiresAt); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to store refresh token", err)
		return
	}

	// Return tokens and user info
	userResponse := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Convert Avatar entity to AvatarDTO if exists
	if user.Avatar != nil {
		userResponse.Avatar = &dto.AvatarDTO{
			ID:        user.Avatar.ID,
			UserID:    user.Avatar.UserID,
			PublicID:  user.Avatar.PublicID,
			PublicURL: user.Avatar.PublicURL,
			SecureURL: user.Avatar.SecureURL,
			CreatedAt: user.Avatar.CreatedAt,
			UpdatedAt: user.Avatar.UpdatedAt,
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "login successful", dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
	})
}

// VerifyEmail handles email verification
// @Summary Verify email address
// @Description Verify user's email address using verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequest true "Verification token"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 410 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err := h.authUseCase.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		if err == errors.ErrInvalidVerificationToken {
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid verification token", err)
			return
		}
		if err == errors.ErrVerificationTokenExpired {
			utils.ErrorResponse(c, http.StatusGone, "verification token expired", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to verify email", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "email verified successfully", nil)
}

// ResendVerification handles resending verification email
// @Summary Resend verification email
// @Description Resend verification email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ResendVerificationRequest true "User email"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req dto.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err := h.authUseCase.ResendVerificationEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err == errors.ErrUserNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "user not found", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to resend verification email", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "verification email sent successfully", nil)
}

// ForgotPassword handles forgot password request
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "User email"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err := h.authUseCase.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to process forgot password request", err)
		return
	}

	// Always return success to prevent email enumeration
	utils.SuccessResponse(c, http.StatusOK, "if the email exists, a password reset link has been sent", nil)
}

// ResetPassword handles password reset
// @Summary Reset password
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 410 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err := h.authUseCase.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		if err == errors.ErrInvalidResetToken {
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid reset token", err)
			return
		}
		if err == errors.ErrResetTokenExpired {
			utils.ErrorResponse(c, http.StatusGone, "reset token expired", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to reset password", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "password reset successfully", nil)
}
