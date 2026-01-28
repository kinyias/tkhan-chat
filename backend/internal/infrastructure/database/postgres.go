package database

import (
	"fmt"

	"backend/internal/infrastructure/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	// Import repository models
	type UserModel struct {
		ID        string `gorm:"primaryKey;type:uuid"`
		Email     string `gorm:"uniqueIndex;not null"`
		Password  string `gorm:"not null"`
		Name      string `gorm:"not null"`
		CreatedAt int64  `gorm:"autoCreateTime:milli"`
		UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	}

	return db.AutoMigrate(&UserModel{})
}
