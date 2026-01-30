package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config holds all configuration for the application
type Config struct {
	// Database
	DatabaseURL string `env:"DATABASE_URL" env-required:"true" env-description:"Database connection URL"`

	// Redis
	RedisURL string `env:"REDIS_URL" env-default:"redis://localhost:6379/0" env-description:"Redis connection URL"`

	// JWT
	JWTSecret string `env:"JWT_SECRET" env-required:"true" env-description:"JWT signing secret"`

	// Centrifugo
	CentrifugoAPIKey   string `env:"CENTRIFUGO_API_KEY" env-required:"true" env-description:"Centrifugo API key"`
	CentrifugoSecret   string `env:"CENTRIFUGO_SECRET" env-required:"true" env-description:"Centrifugo secret"`
	CentrifugoGRPCAddr string `env:"CENTRIFUGO_GRPC_ADDR" env-default:"localhost:8001" env-description:"Centrifugo gRPC address"`

	// TonCenter
	TonCenterAPIKey string `env:"TONCENTER_API_KEY" env-description:"TonCenter API key (required in production)"`

	// Server
	Port        string `env:"PORT" env-default:"8080" env-description:"Server port"`
	MetricsAddr string `env:"METRICS_ADDR" env-default:":9090" env-description:"Metrics server address"`

	// Logging
	LogLevel string `env:"LOG_LEVEL" env-default:"info" env-description:"Log level (debug, info, warn, error)"`

	// Matchmaking
	MatchmakingTimeoutSeconds int `env:"MATCHMAKING_TIMEOUT_SECONDS" env-default:"20" env-description:"Matchmaking timeout in seconds"`

	// Environment
	Environment string `env:"ENVIRONMENT" env-default:"development" env-description:"Application environment (development, production)"`
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	var cfg Config

	// Load configuration from environment variables and .env file
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Additional validation for production-specific requirements
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// validate ensures production-specific configuration requirements are met
func (c *Config) validate() error {
	// TonCenter API key is required in production
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

// Usage prints configuration usage information to stdout
func Usage() {
	var cfg Config
	cleanenv.FUsage(nil, &cfg, nil, nil)
}

