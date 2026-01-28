package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
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
