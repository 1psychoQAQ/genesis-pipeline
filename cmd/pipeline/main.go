package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/filter"
	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
	"github.com/1psychoQAQ/genesis-pipeline/internal/parser/arxiv"
	"github.com/1psychoQAQ/genesis-pipeline/internal/storage"
)

func main() {
	// Command-line flags
	query := flag.String("query", "machine learning", "Search query for ArXiv")
	limit := flag.Int("limit", 10, "Number of papers to fetch")
	minScore := flag.Int("min-score", 60, "Minimum score threshold (0-100)")
	skipDB := flag.Bool("skip-db", false, "Skip database operations")
	skipFilter := flag.Bool("skip-filter", false, "Skip quality filtering")
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

	// Apply quality filtering
	var filteredPapers []model.Paper
	var filterResults []filter.FilterResult
	if *skipFilter {
		filteredPapers = papers
		log.Println("Skipping quality filter (--skip-filter)")
	} else {
		f := filter.NewFilter()
		f.MinScore = *minScore
		filterResults = f.Apply(papers)
		filteredPapers = f.FilterPassed(papers)
		log.Printf("Quality filter: %d/%d papers passed (min score: %d)", len(filteredPapers), len(papers), *minScore)
	}

	// Skip database if requested
	if *skipDB {
		printFilterResults(filterResults, filteredPapers, *skipFilter)
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

	// Save filtered papers
	repo := storage.NewPaperRepository(pool)
	if len(filteredPapers) > 0 {
		if err := repo.SaveBatch(ctx, filteredPapers); err != nil {
			log.Fatalf("Failed to save papers: %v", err)
		}
		log.Printf("Saved %d filtered papers to database", len(filteredPapers))
	} else {
		log.Println("No papers passed the filter, nothing saved")
	}

	// Show count
	count, err := repo.Count(ctx)
	if err != nil {
		log.Fatalf("Failed to count papers: %v", err)
	}
	log.Printf("Total papers in database: %d", count)

	printFilterResults(filterResults, filteredPapers, *skipFilter)
}

func printFilterResults(results []filter.FilterResult, passed []model.Paper, skipFilter bool) {
	fmt.Println("")
	fmt.Println("════════════════════════════════════════════════════════════════")

	if skipFilter {
		// No filter applied, just print papers
		fmt.Printf("  📚 Fetched %d papers (filter skipped):\n", len(passed))
		fmt.Println("════════════════════════════════════════════════════════════════")
		for i, p := range passed {
			fmt.Printf("\n[%d] %s\n", i+1, p.Title)
			fmt.Printf("    Authors: %v\n", p.Authors)
			fmt.Printf("    📄 Abstract: https://arxiv.org/abs/%s\n", p.ID)
			fmt.Printf("    📥 PDF:      https://arxiv.org/pdf/%s.pdf\n", p.ID)
		}
	} else {
		// Only show papers that passed the filter
		fmt.Printf("  📚 Filter Results: %d/%d papers passed\n", len(passed), len(results))
		fmt.Println("════════════════════════════════════════════════════════════════")

		for i, p := range passed {
			fmt.Printf("\n[%d] ✅ %s\n", i+1, p.Title)
			fmt.Printf("    Score: %d/100\n", p.Score)
			if len(p.ScoreDetails) > 0 {
				fmt.Printf("    Details: %s\n", strings.Join(p.ScoreDetails, ", "))
			}
			fmt.Printf("    📄 Abstract: https://arxiv.org/abs/%s\n", p.ID)
			fmt.Printf("    📥 PDF:      https://arxiv.org/pdf/%s.pdf\n", p.ID)
		}
	}

	fmt.Println("")
	fmt.Println("════════════════════════════════════════════════════════════════")
}
