package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	OAuth      OAuthConfig
	Cloudinary CloudinaryConfig
	Email      EmailConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Mode string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret                 string `mapstructure:"secret"`
	AccessTokenExpireMinutes int    `mapstructure:"access_token_expire_minutes"`
	RefreshTokenExpireDays   int    `mapstructure:"refresh_token_expire_days"`
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	GoogleClientID     string `mapstructure:"google_client_id"`
	GoogleClientSecret string `mapstructure:"google_client_secret"`
	GoogleRedirectURL  string `mapstructure:"google_redirect_url"`
}

// CloudinaryConfig holds Cloudinary configuration
type CloudinaryConfig struct {
	CloudName string `mapstructure:"cloud_name"`
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     string `mapstructure:"smtp_port"`
	SMTPUsername string `mapstructure:"smtp_username"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromEmail    string `mapstructure:"from_email"`
	FromName     string `mapstructure:"from_name"`
	FrontendURL  string `mapstructure:"frontend_url"`
}


// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Environment variables override
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables for OAuth
	viper.BindEnv("oauth.google_client_id", "APP_GOOGLE_CLIENT_ID")
	viper.BindEnv("oauth.google_client_secret", "APP_GOOGLE_CLIENT_SECRET")
	viper.BindEnv("oauth.google_redirect_url", "APP_GOOGLE_REDIRECT_URL")

	// Bind specific environment variables for Cloudinary
	viper.BindEnv("cloudinary.cloud_name", "APP_CLOUDINARY_CLOUD_NAME")
	viper.BindEnv("cloudinary.api_key", "APP_CLOUDINARY_API_KEY")
	viper.BindEnv("cloudinary.api_secret", "APP_CLOUDINARY_API_SECRET")

	// Bind specific environment variables for Email
	viper.BindEnv("email.smtp_host", "APP_EMAIL_SMTP_HOST")
	viper.BindEnv("email.smtp_port", "APP_EMAIL_SMTP_PORT")
	viper.BindEnv("email.smtp_username", "APP_EMAIL_SMTP_USERNAME")
	viper.BindEnv("email.smtp_password", "APP_EMAIL_SMTP_PASSWORD")
	viper.BindEnv("email.from_email", "APP_EMAIL_FROM_EMAIL")
	viper.BindEnv("email.from_name", "APP_EMAIL_FROM_NAME")
	viper.BindEnv("email.frontend_url", "APP_EMAIL_FRONTEND_URL")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("jwt.access_token_expire_minutes", 15)
	viper.SetDefault("jwt.refresh_token_expire_days", 7)

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, use defaults and env vars
		fmt.Printf("Config file not found, using defaults and environment variables: %v\n", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
