# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test -v ./internal/parser/arxiv

# Run the pipeline (with database)
go run cmd/pipeline/main.go -query "deep learning" -limit 10

# Run without database
go run cmd/pipeline/main.go -skip-db

# Start PostgreSQL (from project root)
docker-compose -f deployments/docker-compose.yml up -d

# Stop PostgreSQL
docker-compose -f deployments/docker-compose.yml down

# Run benchmark
go run cmd/benchmark/main.go -query "neural networks" -limit 50

# Run Go benchmarks
go test -bench=. ./...
```

## Architecture

Genesis Research Pipeline is a data pipeline for ArXiv scientific literature.

### Project Structure
- `cmd/pipeline/` - Main application with CLI flags
- `cmd/benchmark/` - Benchmark runner
- `internal/model/` - Data models (Paper struct)
- `internal/parser/` - Provider interface for data fetching
- `internal/parser/arxiv/` - ArXiv API client
- `internal/storage/` - PostgreSQL storage layer
- `internal/validation/` - Data quality validation
- `internal/benchmark/` - Benchmark utilities
- `deployments/` - Docker Compose configuration

### Key Interfaces
- `parser.Provider` - Interface for fetching papers: `FetchPapers(query string, limit int) ([]model.Paper, error)`
- `arxiv.Client` - ArXiv API client implementing Provider interface
- `storage.PaperRepository` - CRUD operations for papers (Save, SaveBatch, GetByID, List, Count, Delete)

### Database
PostgreSQL via Docker Compose:
- Host: `localhost:5433`
- Database: `genesis_db`
- User: `genesis`
- Password: `genesis_secret`
