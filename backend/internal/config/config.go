package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret string

	// Centrifugo
	CentrifugoAPIKey   string
	CentrifugoSecret   string
	CentrifugoGRPCAddr string

	// TonCenter
	TonCenterAPIKey string

	// Server
	Port        string
	MetricsAddr string

	// Logging
	LogLevel string

	// Matchmaking
	MatchmakingTimeoutSeconds int

	// Environment
	Environment string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:               getEnv("DATABASE_URL", ""),
		RedisURL:                  getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:                 getEnv("JWT_SECRET", ""),
		CentrifugoAPIKey:          getEnv("CENTRIFUGO_API_KEY", ""),
		CentrifugoSecret:          getEnv("CENTRIFUGO_SECRET", ""),
		CentrifugoGRPCAddr:        getEnv("CENTRIFUGO_GRPC_ADDR", "localhost:8001"),
		TonCenterAPIKey:           getEnv("TONCENTER_API_KEY", ""),
		Port:                      getEnv("PORT", "8080"),
		MetricsAddr:               getEnv("METRICS_ADDR", ":9090"),
		LogLevel:                  getEnv("LOG_LEVEL", "info"),
		MatchmakingTimeoutSeconds: getEnvAsInt("MATCHMAKING_TIMEOUT_SECONDS", 20),
		Environment:               getEnv("ENVIRONMENT", "development"),
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// validate ensures all required configuration is present
func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.CentrifugoAPIKey == "" {
		return fmt.Errorf("CENTRIFUGO_API_KEY is required")
	}
	if c.CentrifugoSecret == "" {
		return fmt.Errorf("CENTRIFUGO_SECRET is required")
	}
	if c.TonCenterAPIKey == "" && c.Environment == "production" {
		return fmt.Errorf("TONCENTER_API_KEY is required in production")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as an integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvAsDuration gets an environment variable as a duration with a fallback value
func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
