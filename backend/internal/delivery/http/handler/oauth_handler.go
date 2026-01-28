package handler

import (
	"net/http"
	"time"

	"backend/internal/delivery/http/dto"
	"backend/internal/usecase/auth"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// OAuthHandler handles HTTP requests for OAuth operations
type OAuthHandler struct {
	oauthUseCase        auth.OAuthUseCase
	jwtService          auth.JWTService
	refreshTokenUseCase auth.RefreshTokenUseCase
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	oauthUseCase auth.OAuthUseCase,
	jwtService auth.JWTService,
	refreshTokenUseCase auth.RefreshTokenUseCase,
) *OAuthHandler {
	return &OAuthHandler{
		oauthUseCase:        oauthUseCase,
		jwtService:          jwtService,
		refreshTokenUseCase: refreshTokenUseCase,
	}
}

// GetGoogleAuthURL generates and returns the Google OAuth authorization URL
// @Summary Get Google OAuth URL
// @Description Get the Google OAuth authorization URL for user login
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} dto.OAuthAuthURLResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/google [get]
func (h *OAuthHandler) GetGoogleAuthURL(c *gin.Context) {
	// Generate state token for CSRF protection
	state, err := h.oauthUseCase.GenerateStateToken()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate state token", err)
		return
	}

	// Store state in session/cookie for validation (in production, use Redis or session store)
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	// Get Google OAuth URL
	authURL := h.oauthUseCase.GetGoogleAuthURL(state)

	utils.SuccessResponse(c, http.StatusOK, "success", dto.OAuthAuthURLResponse{
		AuthURL: authURL,
		State:   state,
	})
}

// HandleGoogleCallback handles the Google OAuth callback
// @Summary Handle Google OAuth callback
// @Description Handle the callback from Google OAuth and authenticate user
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Param state query string true "State token for CSRF protection"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/google/callback [get]
func (h *OAuthHandler) HandleGoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "missing code or state parameter", nil)
		return
	}

	// Validate state token (CSRF protection)
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != state {
		utils.ErrorResponse(c, http.StatusUnauthorized, "invalid state token", nil)
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Handle Google callback
	user, err := h.oauthUseCase.HandleGoogleCallback(c.Request.Context(), code)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to authenticate with Google", err)
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

	utils.SuccessResponse(c, http.StatusOK, "login successful", dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
	})
}
