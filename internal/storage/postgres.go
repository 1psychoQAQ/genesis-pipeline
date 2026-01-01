package storage

import (
	"context"
	"fmt"
	"os"
	"strconv"
)

import "github.com/jackc/pgx/v5/pgxpool"

// Config holds database connection configuration.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// DefaultConfig returns configuration from environment variables or defaults.
func DefaultConfig() Config {
	port := 5433
	if p := os.Getenv("DB_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	return Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     port,
		User:     getEnvOrDefault("DB_USER", "genesis"),
		Password: getEnvOrDefault("DB_PASSWORD", "genesis123"),
		Database: getEnvOrDefault("DB_NAME", "genesis"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// ConnString returns the PostgreSQL connection string.
func (c Config) ConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Database,
	)
}

// NewPool creates a new PostgreSQL connection pool.
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.ConnString())
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
