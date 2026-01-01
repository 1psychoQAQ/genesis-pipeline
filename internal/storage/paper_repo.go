package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
)

// ErrNotFound is returned when a paper is not found.
var ErrNotFound = errors.New("paper not found")

// PaperRepository handles paper persistence.
type PaperRepository struct {
	pool *pgxpool.Pool
}

// NewPaperRepository creates a new paper repository.
func NewPaperRepository(pool *pgxpool.Pool) *PaperRepository {
	return &PaperRepository{pool: pool}
}

// Save inserts or updates a paper.
func (r *PaperRepository) Save(ctx context.Context, paper model.Paper) error {
	query := `
		INSERT INTO papers (id, title, abstract, authors, categories, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			abstract = EXCLUDED.abstract,
			authors = EXCLUDED.authors,
			categories = EXCLUDED.categories,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		paper.ID,
		paper.Title,
		paper.Abstract,
		paper.Authors,
		paper.Categories,
		paper.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save paper: %w", err)
	}

	return nil
}

// SaveBatch inserts or updates multiple papers.
func (r *PaperRepository) SaveBatch(ctx context.Context, papers []model.Paper) error {
	batch := &pgx.Batch{}

	query := `
		INSERT INTO papers (id, title, abstract, authors, categories, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			abstract = EXCLUDED.abstract,
			authors = EXCLUDED.authors,
			categories = EXCLUDED.categories,
			updated_at = EXCLUDED.updated_at
	`

	for _, paper := range papers {
		batch.Queue(query,
			paper.ID,
			paper.Title,
			paper.Abstract,
			paper.Authors,
			paper.Categories,
			paper.UpdatedAt,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range papers {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("batch save: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a paper by ID.
func (r *PaperRepository) GetByID(ctx context.Context, id string) (model.Paper, error) {
	query := `
		SELECT id, title, abstract, authors, categories, updated_at
		FROM papers
		WHERE id = $1
	`

	var paper model.Paper
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&paper.ID,
		&paper.Title,
		&paper.Abstract,
		&paper.Authors,
		&paper.Categories,
		&paper.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Paper{}, ErrNotFound
		}
		return model.Paper{}, fmt.Errorf("get paper: %w", err)
	}

	return paper, nil
}

// List retrieves papers with pagination.
func (r *PaperRepository) List(ctx context.Context, limit, offset int) ([]model.Paper, error) {
	query := `
		SELECT id, title, abstract, authors, categories, updated_at
		FROM papers
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list papers: %w", err)
	}
	defer rows.Close()

	var papers []model.Paper
	for rows.Next() {
		var paper model.Paper
		if err := rows.Scan(
			&paper.ID,
			&paper.Title,
			&paper.Abstract,
			&paper.Authors,
			&paper.Categories,
			&paper.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan paper: %w", err)
		}
		papers = append(papers, paper)
	}

	return papers, nil
}

// Count returns the total number of papers.
func (r *PaperRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM papers").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count papers: %w", err)
	}
	return count, nil
}

// Delete removes a paper by ID.
func (r *PaperRepository) Delete(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM papers WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete paper: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Search searches papers by title or abstract.
func (r *PaperRepository) Search(ctx context.Context, query string, limit int) ([]model.Paper, error) {
	sqlQuery := `
		SELECT id, title, abstract, authors, categories, updated_at
		FROM papers
		WHERE title ILIKE $1 OR abstract ILIKE $1
		ORDER BY updated_at DESC
		LIMIT $2
	`

	searchPattern := "%" + query + "%"
	rows, err := r.pool.Query(ctx, sqlQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search papers: %w", err)
	}
	defer rows.Close()

	var papers []model.Paper
	for rows.Next() {
		var paper model.Paper
		if err := rows.Scan(
			&paper.ID,
			&paper.Title,
			&paper.Abstract,
			&paper.Authors,
			&paper.Categories,
			&paper.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan paper: %w", err)
		}
		papers = append(papers, paper)
	}

	return papers, nil
}

// GetLatestUpdateTime returns the most recent paper update time.
func (r *PaperRepository) GetLatestUpdateTime(ctx context.Context) (time.Time, error) {
	var latest time.Time
	err := r.pool.QueryRow(ctx, "SELECT MAX(updated_at) FROM papers").Scan(&latest)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, ErrNotFound
		}
		return time.Time{}, fmt.Errorf("get latest update: %w", err)
	}
	return latest, nil
}

// Exists checks if a paper with the given ID exists.
func (r *PaperRepository) Exists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM papers WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check exists: %w", err)
	}
	return exists, nil
}

// SaveBatchWithStats saves papers and returns new/updated counts.
func (r *PaperRepository) SaveBatchWithStats(ctx context.Context, papers []model.Paper) (newCount, updatedCount int, err error) {
	for _, paper := range papers {
		exists, err := r.Exists(ctx, paper.ID)
		if err != nil {
			return 0, 0, err
		}

		if err := r.Save(ctx, paper); err != nil {
			return 0, 0, err
		}

		if exists {
			updatedCount++
		} else {
			newCount++
		}
	}
	return newCount, updatedCount, nil
}
