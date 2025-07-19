package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	// Try to load .env file, but don't fail if it doesn't exist
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using environment variables or defaults")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "watson"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func GetBankLinkURL() string {
	return getEnv("BANK_LINK_URL", "http://localhost:5173/")
}

// GetConnectionString returns the PostgreSQL connection string
func (c *Config) GetConnectionString() string {
	return getEnv("DATABASE_URL", "")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
