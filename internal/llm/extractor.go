package llm

// KeywordExtractor defines the interface for extracting search keywords from natural language.
type KeywordExtractor interface {
	// ExtractKeywords takes a natural language question and returns search keywords.
	ExtractKeywords(question string) (string, error)
}

// NewKeywordExtractor creates a keyword extractor based on the provider.
// Supported providers: "gemini" (default)
func NewKeywordExtractor(provider string) (KeywordExtractor, error) {
	switch provider {
	case "gemini", "":
		return NewGeminiClient()
	default:
		return NewGeminiClient()
	}
}
