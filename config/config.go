package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Google OAuth (for YouTube API access)
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string

	// MCP OAuth (for Claude authentication)
	MCPServerURL   string

	// Server config
	Port        string
	Environment string
}

func Load() *Config {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	config := &Config{
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:  getEnv("GOOGLE_REDIRECT_URI", "http://localhost:8080/oauth/callback"),

		MCPServerURL:   getEnv("MCP_SERVER_URL", "http://localhost:8080"),

		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	// Validate required configuration
	if config.GoogleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID environment variable is required")
	}
	if config.GoogleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET environment variable is required")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
