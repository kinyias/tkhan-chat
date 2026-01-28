package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/delivery/http/handler"
	"backend/internal/delivery/http/middleware"
	"backend/internal/delivery/http/router"
	"backend/internal/infrastructure/config"
	"backend/internal/infrastructure/database"
	"backend/internal/infrastructure/logger"
	"backend/internal/repository/postgres"
	"backend/internal/usecase/auth"
	"backend/internal/usecase/user"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Server.Mode)
	defer logger.Sync()

	logger.Info("Starting application...")

	// Connect to database
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", err)
	}

	// Run migrations
	// if err := database.AutoMigrate(db); err != nil {
	// 	logger.Fatal("Failed to run migrations", err)
	// }
	// logger.Info("Database migrations completed")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)

	// Initialize use cases
	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessTokenExpireMinutes, cfg.JWT.RefreshTokenExpireDays)
	userUseCase := user.NewUserUseCase(userRepo)
	refreshTokenUseCase := auth.NewRefreshTokenUseCase(refreshTokenRepo)
	// Initialize OAuth service and use case
	oauthService := auth.NewGoogleOAuthService(cfg.OAuth.GoogleClientID, cfg.OAuth.GoogleClientSecret, cfg.OAuth.GoogleRedirectURL)
	oauthUseCase := auth.NewOAuthUseCase(userRepo, oauthService)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userUseCase, jwtService, refreshTokenUseCase)
	oauthHandler := handler.NewOAuthHandler(oauthUseCase, jwtService, refreshTokenUseCase)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup router
	r := router.NewRouter(userHandler, oauthHandler, authMiddleware)
	ginRouter := r.Setup()

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: ginRouter,
	}

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("Server starting on port %s", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", err)
	}

	logger.Info("Server exited gracefully")
}
