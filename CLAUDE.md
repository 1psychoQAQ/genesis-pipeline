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
go test -v ./internal/parser

# Run the pipeline
go run cmd/pipeline/main.go

# Start PostgreSQL (from project root)
docker-compose -f deployments/docker-compose.yml up -d

# Stop PostgreSQL
docker-compose -f deployments/docker-compose.yml down
```

## Architecture

Genesis Research Pipeline is a data pipeline for ArXiv scientific literature.

### Project Structure
- `cmd/pipeline/` - Application entry point
- `internal/model/` - Data models (Paper struct with ArXiv metadata)
- `internal/parser/` - Data fetching abstractions (Provider interface)
- `deployments/` - Docker Compose configuration for PostgreSQL

### Key Interfaces
- `parser.Provider` - Interface for fetching papers: `FetchPapers(query string, limit int) ([]model.Paper, error)`

### Database
PostgreSQL via Docker Compose:
- Host: `localhost:5432`
- Database: `genesis_db`
- User: `genesis`
- Password: `genesis_secret`
