package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration.
type Config struct {
	// Database configuration
	DB DatabaseConfig

	// Gemini AI configuration
	Gemini GeminiConfig

	// Pipeline defaults
	Pipeline PipelineConfig
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"5433"`
	User     string `envconfig:"DB_USER" default:"genesis"`
	Password string `envconfig:"DB_PASSWORD" default:"genesis123"`
	Name     string `envconfig:"DB_NAME" default:"genesis"`
}

// ConnString returns the PostgreSQL connection string.
func (c DatabaseConfig) ConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Name,
	)
}

// GeminiConfig holds Gemini AI settings.
type GeminiConfig struct {
	APIKey string `envconfig:"GEMINI_API_KEY"`
	Model  string `envconfig:"GEMINI_MODEL" default:"gemini-2.0-flash"`
}

// IsConfigured returns true if API key is set.
func (c GeminiConfig) IsConfigured() bool {
	return c.APIKey != ""
}

// PipelineConfig holds pipeline default settings.
type PipelineConfig struct {
	DefaultQuery    string `envconfig:"DEFAULT_QUERY" default:"machine learning"`
	DefaultLimit    int    `envconfig:"DEFAULT_LIMIT" default:"10"`
	DefaultMinScore int    `envconfig:"DEFAULT_MIN_SCORE" default:"60"`
	DefaultMaxAge   int    `envconfig:"DEFAULT_MAX_AGE" default:"365"`
}

// Load loads configuration from environment variables.
// It first tries to load .env file, then reads environment variables.
func Load() (*Config, error) {
	// Load .env file (optional, won't fail if not exists)
	_ = godotenv.Load()

	var cfg Config

	// Load database config
	if err := envconfig.Process("", &cfg.DB); err != nil {
		return nil, fmt.Errorf("load database config: %w", err)
	}

	// Load Gemini config
	if err := envconfig.Process("", &cfg.Gemini); err != nil {
		return nil, fmt.Errorf("load gemini config: %w", err)
	}

	// Load pipeline config
	if err := envconfig.Process("", &cfg.Pipeline); err != nil {
		return nil, fmt.Errorf("load pipeline config: %w", err)
	}

	return &cfg, nil
}

// MustLoad loads configuration and panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}
