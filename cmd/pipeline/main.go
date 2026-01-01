package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
	"github.com/1psychoQAQ/genesis-pipeline/internal/parser/arxiv"
	"github.com/1psychoQAQ/genesis-pipeline/internal/storage"
)

func main() {
	// Command-line flags
	query := flag.String("query", "machine learning", "Search query for ArXiv")
	limit := flag.Int("limit", 10, "Number of papers to fetch")
	skipDB := flag.Bool("skip-db", false, "Skip database operations")
	flag.Parse()

	log.Println("Genesis Research Pipeline starting...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Fetch papers from ArXiv
	client := arxiv.NewClient()
	log.Printf("Fetching papers for query: %q", *query)

	papers, err := client.FetchPapers(*query, *limit)
	if err != nil {
		log.Fatalf("Failed to fetch papers: %v", err)
	}
	log.Printf("Fetched %d papers from ArXiv", len(papers))

	// Skip database if requested
	if *skipDB {
		printPapers(papers)
		return
	}

	// Connect to database
	cfg := storage.DefaultConfig()
	pool, err := storage.NewPool(ctx, cfg)
	if err != nil {
		log.Printf("Database connection failed: %v", err)
		log.Println("Run with -skip-db flag to skip database operations")
		log.Println("Or start PostgreSQL with: docker-compose -f deployments/docker-compose.yml up -d")
		return
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL")

	// Run migrations
	if err := storage.Migrate(ctx, pool); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migrated")

	// Save papers
	repo := storage.NewPaperRepository(pool)
	if err := repo.SaveBatch(ctx, papers); err != nil {
		log.Fatalf("Failed to save papers: %v", err)
	}
	log.Printf("Saved %d papers to database", len(papers))

	// Show count
	count, err := repo.Count(ctx)
	if err != nil {
		log.Fatalf("Failed to count papers: %v", err)
	}
	log.Printf("Total papers in database: %d", count)

	printPapers(papers)
}

func printPapers(papers []model.Paper) {
	fmt.Println("")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Printf("  📚 Fetched %d papers:\n", len(papers))
	fmt.Println("════════════════════════════════════════════════════════════════")
	for i, p := range papers {
		fmt.Printf("\n[%d] %s\n", i+1, p.Title)
		fmt.Printf("    Authors: %v\n", p.Authors)
		fmt.Printf("    Categories: %v\n", p.Categories)
		fmt.Printf("    📄 Abstract: https://arxiv.org/abs/%s\n", p.ID)
		fmt.Printf("    📥 PDF:      https://arxiv.org/pdf/%s.pdf\n", p.ID)
	}
	fmt.Println("")
	fmt.Println("════════════════════════════════════════════════════════════════")
}
