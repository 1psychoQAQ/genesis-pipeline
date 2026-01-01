package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS papers (
    id VARCHAR(50) PRIMARY KEY,
    title TEXT NOT NULL,
    abstract TEXT NOT NULL,
    authors TEXT[] NOT NULL DEFAULT '{}',
    categories TEXT[] NOT NULL DEFAULT '{}',
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_papers_updated_at ON papers(updated_at);
CREATE INDEX IF NOT EXISTS idx_papers_categories ON papers USING GIN(categories);

CREATE TABLE IF NOT EXISTS sync_log (
    id SERIAL PRIMARY KEY,
    query VARCHAR(255) NOT NULL,
    papers_fetched INT NOT NULL DEFAULT 0,
    papers_new INT NOT NULL DEFAULT 0,
    papers_updated INT NOT NULL DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'running'
);
`

// Migrate runs database migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}
	return nil
}
