package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncLog represents a synchronization operation log.
type SyncLog struct {
	ID            int
	Query         string
	PapersFetched int
	PapersNew     int
	PapersUpdated int
	StartedAt     time.Time
	CompletedAt   *time.Time
	Status        string
}

// SyncRepository handles sync log persistence.
type SyncRepository struct {
	pool *pgxpool.Pool
}

// NewSyncRepository creates a new sync repository.
func NewSyncRepository(pool *pgxpool.Pool) *SyncRepository {
	return &SyncRepository{pool: pool}
}

// StartSync creates a new sync log entry and returns its ID.
func (r *SyncRepository) StartSync(ctx context.Context, query string) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO sync_log (query, started_at, status)
		VALUES ($1, NOW(), 'running')
		RETURNING id
	`, query).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("start sync: %w", err)
	}
	return id, nil
}

// CompleteSync updates a sync log entry with results.
func (r *SyncRepository) CompleteSync(ctx context.Context, id int, fetched, newCount, updated int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sync_log
		SET papers_fetched = $2,
		    papers_new = $3,
		    papers_updated = $4,
		    completed_at = NOW(),
		    status = 'completed'
		WHERE id = $1
	`, id, fetched, newCount, updated)
	if err != nil {
		return fmt.Errorf("complete sync: %w", err)
	}
	return nil
}

// FailSync marks a sync as failed.
func (r *SyncRepository) FailSync(ctx context.Context, id int, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sync_log
		SET completed_at = NOW(),
		    status = 'failed'
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("fail sync: %w", err)
	}
	return nil
}

// GetLatestSync returns the most recent completed sync.
func (r *SyncRepository) GetLatestSync(ctx context.Context) (*SyncLog, error) {
	var log SyncLog
	err := r.pool.QueryRow(ctx, `
		SELECT id, query, papers_fetched, papers_new, papers_updated,
		       started_at, completed_at, status
		FROM sync_log
		WHERE status = 'completed'
		ORDER BY completed_at DESC
		LIMIT 1
	`).Scan(
		&log.ID, &log.Query, &log.PapersFetched, &log.PapersNew,
		&log.PapersUpdated, &log.StartedAt, &log.CompletedAt, &log.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("get latest sync: %w", err)
	}
	return &log, nil
}

// GetSyncHistory returns recent sync operations.
func (r *SyncRepository) GetSyncHistory(ctx context.Context, limit int) ([]SyncLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, query, papers_fetched, papers_new, papers_updated,
		       started_at, completed_at, status
		FROM sync_log
		ORDER BY started_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("get sync history: %w", err)
	}
	defer rows.Close()

	var logs []SyncLog
	for rows.Next() {
		var log SyncLog
		if err := rows.Scan(
			&log.ID, &log.Query, &log.PapersFetched, &log.PapersNew,
			&log.PapersUpdated, &log.StartedAt, &log.CompletedAt, &log.Status,
		); err != nil {
			return nil, fmt.Errorf("scan sync log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}
