package arxiv

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
)

const (
	defaultBaseURL = "http://export.arxiv.org/api/query"
	defaultTimeout = 30 * time.Second
)

// Client is an ArXiv API client that implements the parser.Provider interface.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new ArXiv API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: defaultBaseURL,
	}
}

// NewClientWithOptions creates a new ArXiv API client with custom options.
func NewClientWithOptions(httpClient *http.Client, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// FetchPapers retrieves papers from ArXiv matching the query.
func (c *Client) FetchPapers(query string, limit int) ([]model.Paper, error) {
	if limit <= 0 {
		limit = 10
	}

	reqURL, err := c.buildURL(query, limit)
	if err != nil {
		return nil, fmt.Errorf("build URL: %w", err)
	}

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var feed atomFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decode XML: %w", err)
	}

	return c.convertEntries(feed.Entries), nil
}

func (c *Client) buildURL(query string, limit int) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("search_query", fmt.Sprintf("all:%s", query))
	q.Set("start", "0")
	q.Set("max_results", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c *Client) convertEntries(entries []atomEntry) []model.Paper {
	papers := make([]model.Paper, 0, len(entries))

	for _, entry := range entries {
		paper := model.Paper{
			ID:         extractID(entry.ID),
			Title:      cleanText(entry.Title),
			Abstract:   cleanText(entry.Summary),
			Authors:    extractAuthors(entry.Authors),
			Categories: extractCategories(entry.Categories),
			UpdatedAt:  entry.Updated,
			Comments:   cleanText(entry.Comment),
			DOI:        strings.TrimSpace(entry.DOI),
			JournalRef: strings.TrimSpace(entry.JournalRef),
			Links:      extractLinks(entry.Links),
		}
		papers = append(papers, paper)
	}

	return papers
}

func extractLinks(links []atomLink) []model.Link {
	result := make([]model.Link, 0, len(links))
	for _, l := range links {
		if l.Href == "" {
			continue
		}
		linkType := "other"
		if strings.Contains(l.Type, "pdf") || strings.HasSuffix(l.Href, ".pdf") {
			linkType = "pdf"
		} else if l.Rel == "alternate" {
			linkType = "abstract"
		} else if strings.Contains(l.Href, "github.com") || strings.Contains(l.Href, "gitlab.com") {
			linkType = "code"
		}
		result = append(result, model.Link{
			URL:   l.Href,
			Type:  linkType,
			Title: l.Title,
		})
	}
	return result
}

// extractID extracts the paper ID from the ArXiv URL.
// Example: "http://arxiv.org/abs/2301.00001v1" -> "2301.00001v1"
func extractID(rawID string) string {
	parts := strings.Split(rawID, "/abs/")
	if len(parts) == 2 {
		return parts[1]
	}
	return rawID
}

func extractAuthors(authors []atomAuthor) []string {
	names := make([]string, 0, len(authors))
	for _, a := range authors {
		if name := strings.TrimSpace(a.Name); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func extractCategories(categories []atomCategory) []string {
	terms := make([]string, 0, len(categories))
	for _, c := range categories {
		if term := strings.TrimSpace(c.Term); term != "" {
			terms = append(terms, term)
		}
	}
	return terms
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	// Collapse multiple spaces into one
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}
