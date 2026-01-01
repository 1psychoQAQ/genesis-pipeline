package parser

import "github.com/1psychoQAQ/genesis-pipeline/internal/model"

// Provider defines the interface for fetching papers from external sources.
type Provider interface {
	// FetchPapers retrieves papers matching the query, up to the specified limit.
	FetchPapers(query string, limit int) ([]model.Paper, error)
}
