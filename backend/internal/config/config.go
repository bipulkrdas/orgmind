package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	ServerPort string

	// Database
	DatabaseURL string

	// JWT
	JWTSecret          string
	JWTExpirationHours int

	// AWS S3
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSS3Bucket        string

	// Zep Cloud
	ZepAPIKey string
	ZepAPIURL string

	// OAuth - Google
	GoogleClientID     string
	GoogleClientSecret string

	// OAuth - Okta
	OktaDomain       string
	OktaClientID     string
	OktaClientSecret string

	// OAuth - Office365
	Office365ClientID     string
	Office365ClientSecret string

	// OAuth Redirect
	OAuthRedirectURL string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file from backend directory (optional in production)
	// This works whether you run from project root or backend directory
	loadEnvFile()

	cfg := &Config{
		ServerPort:            getEnv("SERVER_PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		JWTExpirationHours:    getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		AWSRegion:             getEnv("AWS_REGION", ""),
		AWSAccessKeyID:        getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:    getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSS3Bucket:           getEnv("AWS_S3_BUCKET", ""),
		ZepAPIKey:             getEnv("ZEP_API_KEY", ""),
		ZepAPIURL:             getEnv("ZEP_API_URL", "https://api.getzep.com/api/v2"),
		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		OktaDomain:            getEnv("OKTA_DOMAIN", ""),
		OktaClientID:          getEnv("OKTA_CLIENT_ID", ""),
		OktaClientSecret:      getEnv("OKTA_CLIENT_SECRET", ""),
		Office365ClientID:     getEnv("OFFICE365_CLIENT_ID", ""),
		Office365ClientSecret: getEnv("OFFICE365_CLIENT_SECRET", ""),
		OAuthRedirectURL:      getEnv("OAUTH_REDIRECT_URL", ""),
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that all required configuration values are present
func (c *Config) Validate() error {
	required := map[string]string{
		"DATABASE_URL": c.DatabaseURL,
		"JWT_SECRET":   c.JWTSecret,
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// loadEnvFile attempts to load .env file from multiple possible locations
// This ensures it works whether running from project root or backend directory
func loadEnvFile() {
	// Try multiple possible locations for .env file
	possiblePaths := []string{
		".env",                           // Current directory
		"backend/.env",                   // From project root
		"../.env",                        // From cmd/server directory
		"../../.env",                     // From internal/config directory
		filepath.Join("backend", ".env"), // Explicit backend path
	}

	for _, path := range possiblePaths {
		if err := godotenv.Load(path); err == nil {
			// Successfully loaded .env file
			return
		}
	}

	// If no .env file found, that's okay - we'll use environment variables
	// This is expected in production environments
}
