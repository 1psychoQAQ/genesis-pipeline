package llm

import "github.com/1psychoQAQ/genesis-pipeline/internal/config"

// KeywordExtractor defines the interface for extracting search keywords from natural language.
type KeywordExtractor interface {
	// ExtractKeywords takes a natural language question and returns search keywords.
	ExtractKeywords(question string) (string, error)
}

// NewKeywordExtractor creates a keyword extractor based on the provider.
// Supported providers: "gemini" (default)
func NewKeywordExtractor(provider string, cfg config.GeminiConfig) (KeywordExtractor, error) {
	switch provider {
	case "gemini", "":
		return NewGeminiClient(cfg)
	default:
		return NewGeminiClient(cfg)
	}
}
