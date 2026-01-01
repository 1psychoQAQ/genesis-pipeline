package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/benchmark"
	"github.com/1psychoQAQ/genesis-pipeline/internal/parser/arxiv"
)

func main() {
	query := flag.String("query", "machine learning", "Search query for ArXiv")
	limit := flag.Int("limit", 50, "Number of papers to fetch")
	flag.Parse()

	log.Println("Starting benchmark...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client := arxiv.NewClient()
	runner := benchmark.NewRunner(client)

	report, err := runner.GenerateReport(ctx, *query, *limit)
	if err != nil {
		log.Fatalf("Benchmark failed: %v", err)
	}

	benchmark.PrintReport(report)
}
