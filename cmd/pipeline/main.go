package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/config"
	"github.com/1psychoQAQ/genesis-pipeline/internal/filter"
	"github.com/1psychoQAQ/genesis-pipeline/internal/llm"
	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
	"github.com/1psychoQAQ/genesis-pipeline/internal/parser/arxiv"
	"github.com/1psychoQAQ/genesis-pipeline/internal/storage"
)

func main() {
	// Load configuration from .env and environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Command-line flags (override config defaults)
	question := flag.String("question", "", "Natural language question (uses AI to extract keywords)")
	query := flag.String("query", "", "Direct search query for ArXiv")
	limit := flag.Int("limit", cfg.Pipeline.DefaultLimit, "Number of papers to fetch")
	minScore := flag.Int("min-score", cfg.Pipeline.DefaultMinScore, "Minimum score threshold (0-100)")
	maxAgeDays := flag.Int("max-age", cfg.Pipeline.DefaultMaxAge, "Maximum paper age in days (0 = no limit)")
	skipDB := flag.Bool("skip-db", false, "Skip database operations")
	skipFilter := flag.Bool("skip-filter", false, "Skip quality filtering")
	flag.Parse()

	log.Println("Genesis Research Pipeline starting...")

	// Determine search query
	searchQuery := *query
	if *question != "" {
		// Use LLM to extract keywords from question
		if !cfg.Gemini.IsConfigured() {
			log.Fatalf("GEMINI_API_KEY not configured. Please set it in .env file")
		}

		log.Printf("Processing question: %q", *question)
		extractor, err := llm.NewKeywordExtractor("gemini", cfg.Gemini)
		if err != nil {
			log.Fatalf("Failed to create keyword extractor: %v", err)
		}

		keywords, err := extractor.ExtractKeywords(*question)
		if err != nil {
			log.Fatalf("Failed to extract keywords: %v", err)
		}
		searchQuery = keywords
		log.Printf("AI extracted keywords: %q", searchQuery)
	} else if searchQuery == "" {
		searchQuery = cfg.Pipeline.DefaultQuery
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Fetch papers from ArXiv
	client := arxiv.NewClient()
	log.Printf("Fetching papers for query: %q", searchQuery)

	papers, err := client.FetchPapers(searchQuery, *limit)
	if err != nil {
		log.Fatalf("Failed to fetch papers: %v", err)
	}
	log.Printf("Fetched %d papers from ArXiv", len(papers))

	// Apply time filter (recency)
	if *maxAgeDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -*maxAgeDays)
		var recentPapers []model.Paper
		for _, p := range papers {
			if p.UpdatedAt.After(cutoff) {
				recentPapers = append(recentPapers, p)
			}
		}
		log.Printf("Time filter: %d/%d papers within %d days", len(recentPapers), len(papers), *maxAgeDays)
		papers = recentPapers
	}

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
	pool, err := storage.NewPool(ctx, cfg.DB)
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
			fmt.Printf("    Score: %d/100 | Updated: %s\n", p.Score, p.UpdatedAt.Format("2006-01-02"))
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
