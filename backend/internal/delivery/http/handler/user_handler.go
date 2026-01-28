package handler

import (
	"net/http"
	"strconv"
	"time"

	"backend/internal/delivery/http/dto"
	"backend/internal/domain/entity"
	"backend/internal/usecase/auth"
	"backend/internal/usecase/user"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userUseCase         user.UserUseCase
	jwtService          auth.JWTService
	refreshTokenUseCase auth.RefreshTokenUseCase
	validate            *validator.Validate
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUseCase user.UserUseCase, jwtService auth.JWTService, refreshTokenUseCase auth.RefreshTokenUseCase) *UserHandler {
	return &UserHandler{
		userUseCase:         userUseCase,
		jwtService:          jwtService,
		refreshTokenUseCase: refreshTokenUseCase,
		validate:            validator.New(),
	}
}

// Register handles user registration
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	user, err := h.userUseCase.Register(c.Request.Context(), req.Email, req.Password, req.Name, req.Phone)
	if err != nil {
		utils.HandleDomainError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "user registered successfully", h.toUserResponse(user))
}

// Login handles user authentication
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	user, err := h.userUseCase.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		utils.HandleDomainError(c, err)
		return
	}

	// Generate access token
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate access token", err)
		return
	}

	// Generate refresh token
	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate refresh token", err)
		return
	}

	// Store refresh token in database
	expiresAt := time.Now().Add(h.jwtService.GetRefreshTokenExpiration())
	if err := h.refreshTokenUseCase.CreateRefreshToken(c.Request.Context(), user.ID, refreshToken, expiresAt); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to store refresh token", err)
		return
	}

	response := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         h.toUserResponse(user),
	}

	utils.SuccessResponse(c, http.StatusOK, "login successful", response)
}

// RefreshToken handles token refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate JWT refresh token
	claims, err := h.jwtService.ValidateToken(req.RefreshToken, auth.RefreshToken)
	if err != nil {
		utils.HandleDomainError(c, err)
		return
	}

	// Validate refresh token in database
	storedToken, err := h.refreshTokenUseCase.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.HandleDomainError(c, err)
		return
	}

	// Verify token belongs to the user
	if storedToken.UserID != claims.UserID {
		utils.ErrorResponse(c, http.StatusUnauthorized, "invalid token", nil)
		return
	}

	// Generate new access token
	newAccessToken, err := h.jwtService.GenerateAccessToken(claims.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate access token", err)
		return
	}

	// Generate new refresh token
	newRefreshToken, err := h.jwtService.GenerateRefreshToken(claims.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate refresh token", err)
		return
	}

	// Revoke old refresh token
	if err := h.refreshTokenUseCase.RevokeRefreshToken(c.Request.Context(), req.RefreshToken); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to revoke old token", err)
		return
	}

	// Store new refresh token
	expiresAt := time.Now().Add(h.jwtService.GetRefreshTokenExpiration())
	if err := h.refreshTokenUseCase.CreateRefreshToken(c.Request.Context(), claims.UserID, newRefreshToken, expiresAt); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to store refresh token", err)
		return
	}

	response := &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	utils.SuccessResponse(c, http.StatusOK, "token refreshed successfully", response)
}

// Logout handles user logout by revoking all refresh tokens
func (h *UserHandler) Logout(c *gin.Context) {
	userID := c.GetString("userID")

	if err := h.refreshTokenUseCase.RevokeAllUserTokens(c.Request.Context(), userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to logout", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "logout successful", nil)
}

// GetProfile retrieves the authenticated user's profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")

	user, err := h.userUseCase.GetByID(c.Request.Context(), userID)
	if err != nil {
		utils.HandleDomainError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "profile retrieved successfully", h.toUserResponse(user))
}

// UpdateProfile updates the authenticated user's profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")
	var req dto.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	user, err := h.userUseCase.Update(c.Request.Context(), userID, req.Name, req.Phone)
	if err != nil {
		utils.HandleDomainError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "profile updated successfully", h.toUserResponse(user))
}

// GetUserByID retrieves a user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.HandleDomainError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "user retrieved successfully", h.toUserResponse(user))
}

// ListUsers retrieves a list of users with pagination
func (h *UserHandler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.userUseCase.List(c.Request.Context(), limit, offset)
	if err != nil {
		utils.HandleDomainError(c, err)
		return
	}

	response := &dto.ListUsersResponse{
		Users:  h.toUserResponseList(users),
		Total:  len(users),
		Limit:  limit,
		Offset: offset,
	}

	utils.SuccessResponse(c, http.StatusOK, "users retrieved successfully", response)
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.userUseCase.Delete(c.Request.Context(), id); err != nil {
		utils.HandleDomainError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "user deleted successfully", nil)
}

// UpdateAvatar handles avatar upload
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID := c.GetString("userID")

	// Get file from request
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "avatar file is required", err)
		return
	}
	defer file.Close()

	// Validate file size (max 5MB)
	const maxFileSize = 5 * 1024 * 1024 // 5MB
	if header.Size > maxFileSize {
		utils.ErrorResponse(c, http.StatusBadRequest, "file size exceeds 5MB limit", nil)
		return
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid file type. Allowed: jpeg, jpg, png, gif, webp", nil)
		return
	}

	// Update avatar
	user, err := h.userUseCase.UpdateAvatar(c.Request.Context(), userID, file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "avatar updated successfully", h.toUserResponse(user))
}

// toUserResponse converts entity to response DTO
func (h *UserHandler) toUserResponse(user *entity.User) *dto.UserResponse {
	response := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Convert Avatar entity to AvatarDTO if exists
	if user.Avatar != nil {
		response.Avatar = &dto.AvatarDTO{
			ID:        user.Avatar.ID,
			UserID:    user.Avatar.UserID,
			PublicID:  user.Avatar.PublicID,
			PublicURL: user.Avatar.PublicURL,
			SecureURL: user.Avatar.SecureURL,
			CreatedAt: user.Avatar.CreatedAt,
			UpdatedAt: user.Avatar.UpdatedAt,
		}
	}

	return response
}

// toUserResponseList converts entity list to response DTO list
func (h *UserHandler) toUserResponseList(users []*entity.User) []*dto.UserResponse {
	responses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		responses[i] = h.toUserResponse(user)
	}
	return responses
}
