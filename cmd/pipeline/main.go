package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/filter"
	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
	"github.com/1psychoQAQ/genesis-pipeline/internal/parser/arxiv"
	"github.com/1psychoQAQ/genesis-pipeline/internal/preset"
	"github.com/1psychoQAQ/genesis-pipeline/internal/storage"
)

func main() {
	// Command-line flags
	query := flag.String("query", "", "Search query for ArXiv")
	presetName := flag.String("preset", "", "Use a search preset (use -list-presets to see options)")
	listPresets := flag.Bool("list-presets", false, "List all available search presets")
	limit := flag.Int("limit", 50, "Number of papers to fetch")
	minScore := flag.Int("min-score", 0, "Minimum score threshold (0-100, 0 = use preset default)")
	maxAgeDays := flag.Int("max-age", 0, "Maximum paper age in days (0 = use preset default)")
	skipDB := flag.Bool("skip-db", false, "Skip database operations")
	skipFilter := flag.Bool("skip-filter", false, "Skip quality filtering")
	flag.Parse()

	// List presets and exit
	if *listPresets {
		printPresets()
		return
	}

	// Determine search parameters
	searchQuery := *query
	effectiveMinScore := *minScore
	effectiveMaxAge := *maxAgeDays

	if *presetName != "" {
		p, ok := preset.Get(*presetName)
		if !ok {
			log.Fatalf("Unknown preset: %q. Use -list-presets to see available options.", *presetName)
		}
		searchQuery = p.Query
		if effectiveMinScore == 0 {
			effectiveMinScore = p.MinScore
		}
		if effectiveMaxAge == 0 {
			effectiveMaxAge = p.MaxAgeDays
		}
		log.Printf("Using preset: %s (%s)", p.Name, p.Description)
	}

	// Default values if not using preset
	if searchQuery == "" {
		searchQuery = "machine learning"
	}
	if effectiveMinScore == 0 {
		effectiveMinScore = 60
	}
	if effectiveMaxAge == 0 {
		effectiveMaxAge = 365
	}

	log.Println("Genesis Research Pipeline starting...")

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
	if effectiveMaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -effectiveMaxAge)
		var recentPapers []model.Paper
		for _, p := range papers {
			if p.UpdatedAt.After(cutoff) {
				recentPapers = append(recentPapers, p)
			}
		}
		log.Printf("Time filter: %d/%d papers within %d days", len(recentPapers), len(papers), effectiveMaxAge)
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
		f.MinScore = effectiveMinScore
		filterResults = f.Apply(papers)
		filteredPapers = f.FilterPassed(papers)
		log.Printf("Quality filter: %d/%d papers passed (min score: %d)", len(filteredPapers), len(papers), effectiveMinScore)
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

func printPresets() {
	fmt.Println("")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("  Available Search Presets")
	fmt.Println("════════════════════════════════════════════════════════════════")

	// Group presets by category
	categories := map[string][]string{
		"LLM & NLP":         {"llm-reasoning", "llm-agent", "llm-eval", "rag", "prompt"},
		"Computer Vision":   {"diffusion", "multimodal", "video"},
		"Machine Learning":  {"transformer", "finetune", "distill", "rl"},
		"Safety & Alignment": {"alignment", "jailbreak", "hallucination"},
		"Data & Training":   {"data-synthesis", "scaling"},
	}

	categoryOrder := []string{"LLM & NLP", "Computer Vision", "Machine Learning", "Safety & Alignment", "Data & Training"}

	for _, cat := range categoryOrder {
		names := categories[cat]
		fmt.Printf("\n  [%s]\n", cat)
		for _, name := range names {
			p, _ := preset.Get(name)
			fmt.Printf("    %-18s %s\n", p.Name, p.Description)
		}
	}

	fmt.Println("")
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Println("  Usage: go run cmd/pipeline/main.go -preset <name> [-limit N]")
	fmt.Println("  Example: go run cmd/pipeline/main.go -preset llm-reasoning")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("")

	// Also print all presets sorted for reference
	fmt.Println("All presets with details:")
	fmt.Println("")

	presetList := preset.List()
	sort.Slice(presetList, func(i, j int) bool {
		return presetList[i].Name < presetList[j].Name
	})

	for _, p := range presetList {
		fmt.Printf("  %s:\n", p.Name)
		fmt.Printf("    Query: %s\n", p.Query)
		fmt.Printf("    MinScore: %d, MaxAge: %d days\n", p.MinScore, p.MaxAgeDays)
		fmt.Println("")
	}
}
