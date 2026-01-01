package main

import (
	"fmt"
	"log"

	"github.com/1psychoQAQ/genesis-pipeline/internal/parser/arxiv"
)

func main() {
	log.Println("Genesis Research Pipeline starting...")

	client := arxiv.NewClient()

	papers, err := client.FetchPapers("machine learning", 5)
	if err != nil {
		log.Fatalf("Failed to fetch papers: %v", err)
	}

	fmt.Printf("Fetched %d papers:\n\n", len(papers))
	for i, paper := range papers {
		fmt.Printf("[%d] %s\n", i+1, paper.Title)
		fmt.Printf("    ID: %s\n", paper.ID)
		fmt.Printf("    Authors: %v\n", paper.Authors)
		fmt.Printf("    Categories: %v\n", paper.Categories)
		fmt.Printf("    Updated: %s\n\n", paper.UpdatedAt.Format("2006-01-02"))
	}
}
