package router

import (
	"backend/internal/delivery/http/handler"
	"backend/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
)

// Router manages all HTTP routes
type Router struct {
	userHandler    *handler.UserHandler
	oauthHandler   *handler.OAuthHandler
	authMiddleware *middleware.AuthMiddleware
}

// NewRouter creates a new router
func NewRouter(
	userHandler *handler.UserHandler,
	oauthHandler *handler.OAuthHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		userHandler:    userHandler,
		oauthHandler:   oauthHandler,
		authMiddleware: authMiddleware,
	}
}

// Setup configures all routes
func (r *Router) Setup() *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Public routes - Authentication
		auth := v1.Group("/auth")
		{
			auth.POST("/register", r.userHandler.Register)
			auth.POST("/login", r.userHandler.Login)
			auth.POST("/refresh", r.userHandler.RefreshToken)
			
			// Google OAuth routes
			auth.GET("/google", r.oauthHandler.GetGoogleAuthURL)
			auth.GET("/google/callback", r.oauthHandler.HandleGoogleCallback)
		}

		// Protected auth routes
		authProtected := v1.Group("/auth")
		authProtected.Use(r.authMiddleware.Authenticate())
		{
			authProtected.POST("/logout", r.userHandler.Logout)
		}

		// Protected routes - User profile
		users := v1.Group("/users")
		users.Use(r.authMiddleware.Authenticate())
		{
			users.GET("/me", r.userHandler.GetProfile)
			users.PUT("/me", r.userHandler.UpdateProfile)
			users.GET("/:id", r.userHandler.GetUserByID)
			users.GET("", r.userHandler.ListUsers)
			users.DELETE("/:id", r.userHandler.DeleteUser)
		}
	}

	return router
}

