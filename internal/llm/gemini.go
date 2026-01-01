package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	geminiAPIBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"
	defaultModel     = "gemini-2.0-flash"
)

// GeminiClient handles Gemini API calls and implements KeywordExtractor.
type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGeminiClient creates a new Gemini client.
// Configuration from environment variables:
//   - GEMINI_API_KEY: API key (required)
//   - GEMINI_MODEL: Model name (optional, default: gemini-2.0-flash)
func NewGeminiClient() (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = defaultModel
	}

	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// geminiRequest represents the API request structure.
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

// geminiResponse represents the API response structure.
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ExtractKeywords takes a natural language question and returns search keywords.
func (c *GeminiClient) ExtractKeywords(question string) (string, error) {
	prompt := fmt.Sprintf(`You are a research assistant. Given a research question, extract the most relevant English keywords for searching academic papers on ArXiv.

Rules:
1. Output ONLY the keywords, separated by spaces
2. Use 3-6 keywords maximum
3. Use technical/academic terms
4. Keywords must be in English
5. Do not include common words like "how", "what", "why"
6. Focus on the core concepts and methods

Question: %s

Keywords:`, question)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiAPIBaseURL, c.model, c.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request: %w", err)
	}
	defer resp.Body.Close()

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	keywords := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	return keywords, nil
}

// Model returns the current model name.
func (c *GeminiClient) Model() string {
	return c.model
}
