package benchmark

import (
	"context"
	"fmt"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
	"github.com/1psychoQAQ/genesis-pipeline/internal/parser"
	"github.com/1psychoQAQ/genesis-pipeline/internal/validation"
)

// Result holds benchmark results.
type Result struct {
	Operation     string
	Duration      time.Duration
	ItemCount     int
	ItemsPerSec   float64
	ValidationRes *validation.ValidationResult
}

func (r Result) String() string {
	return fmt.Sprintf(
		"%s: %v (%d items, %.2f items/sec)",
		r.Operation, r.Duration, r.ItemCount, r.ItemsPerSec,
	)
}

// Runner executes benchmarks on the pipeline.
type Runner struct {
	provider parser.Provider
}

// NewRunner creates a new benchmark runner.
func NewRunner(provider parser.Provider) *Runner {
	return &Runner{provider: provider}
}

// BenchmarkFetch measures paper fetching performance.
func (r *Runner) BenchmarkFetch(ctx context.Context, query string, limit int) (Result, []model.Paper, error) {
	start := time.Now()

	papers, err := r.provider.FetchPapers(query, limit)
	if err != nil {
		return Result{}, nil, err
	}

	duration := time.Since(start)
	itemsPerSec := float64(len(papers)) / duration.Seconds()

	// Validate fetched papers
	valResult := validation.ValidatePapers(papers)

	return Result{
		Operation:     "Fetch",
		Duration:      duration,
		ItemCount:     len(papers),
		ItemsPerSec:   itemsPerSec,
		ValidationRes: &valResult,
	}, papers, nil
}

// BenchmarkValidation measures validation performance.
func (r *Runner) BenchmarkValidation(papers []model.Paper) Result {
	start := time.Now()

	valResult := validation.ValidatePapers(papers)

	duration := time.Since(start)
	itemsPerSec := float64(len(papers)) / duration.Seconds()

	return Result{
		Operation:     "Validation",
		Duration:      duration,
		ItemCount:     len(papers),
		ItemsPerSec:   itemsPerSec,
		ValidationRes: &valResult,
	}
}

// Report holds a complete benchmark report.
type Report struct {
	Timestamp time.Time
	Results   []Result
	Summary   Summary
}

// Summary holds summary statistics.
type Summary struct {
	TotalPapers   int
	ValidPapers   int
	InvalidPapers int
	TotalDuration time.Duration
}

// GenerateReport runs all benchmarks and generates a report.
func (r *Runner) GenerateReport(ctx context.Context, query string, limit int) (*Report, error) {
	report := &Report{
		Timestamp: time.Now(),
	}

	// Benchmark fetch
	fetchResult, papers, err := r.BenchmarkFetch(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch benchmark: %w", err)
	}
	report.Results = append(report.Results, fetchResult)

	// Benchmark validation
	valResult := r.BenchmarkValidation(papers)
	report.Results = append(report.Results, valResult)

	// Calculate summary
	var totalDuration time.Duration
	for _, res := range report.Results {
		totalDuration += res.Duration
	}

	report.Summary = Summary{
		TotalPapers:   len(papers),
		ValidPapers:   fetchResult.ValidationRes.Valid,
		InvalidPapers: fetchResult.ValidationRes.Invalid,
		TotalDuration: totalDuration,
	}

	return report, nil
}

// PrintReport prints a benchmark report to stdout.
func PrintReport(report *Report) {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("         BENCHMARK REPORT")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Printf("Timestamp: %s\n\n", report.Timestamp.Format(time.RFC3339))

	fmt.Println("Results:")
	fmt.Println("───────────────────────────────────────────")
	for _, r := range report.Results {
		fmt.Printf("  %s\n", r)
		if r.ValidationRes != nil {
			fmt.Printf("    Valid: %d, Invalid: %d\n",
				r.ValidationRes.Valid, r.ValidationRes.Invalid)
		}
	}

	fmt.Println("\nSummary:")
	fmt.Println("───────────────────────────────────────────")
	fmt.Printf("  Total Papers:   %d\n", report.Summary.TotalPapers)
	fmt.Printf("  Valid Papers:   %d\n", report.Summary.ValidPapers)
	fmt.Printf("  Invalid Papers: %d\n", report.Summary.InvalidPapers)
	fmt.Printf("  Total Duration: %v\n", report.Summary.TotalDuration)
	fmt.Println("═══════════════════════════════════════════")
}
