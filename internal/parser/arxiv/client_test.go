package arxiv

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const mockResponse = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <id>http://arxiv.org/abs/2301.00001v1</id>
    <title>Test Paper Title</title>
    <summary>This is the abstract of the test paper.
    It spans multiple lines.</summary>
    <updated>2023-01-15T10:00:00Z</updated>
    <published>2023-01-01T00:00:00Z</published>
    <author>
      <name>John Doe</name>
    </author>
    <author>
      <name>Jane Smith</name>
    </author>
    <category term="cs.AI" />
    <category term="cs.LG" />
  </entry>
</feed>`

func TestClient_FetchPapers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		query := r.URL.Query()
		if query.Get("search_query") == "" {
			t.Error("expected search_query parameter")
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClientWithOptions(server.Client(), server.URL)

	papers, err := client.FetchPapers("machine learning", 10)
	if err != nil {
		t.Fatalf("FetchPapers failed: %v", err)
	}

	if len(papers) != 1 {
		t.Fatalf("expected 1 paper, got %d", len(papers))
	}

	paper := papers[0]
	if paper.ID != "2301.00001v1" {
		t.Errorf("expected ID '2301.00001v1', got %q", paper.ID)
	}
	if paper.Title != "Test Paper Title" {
		t.Errorf("expected title 'Test Paper Title', got %q", paper.Title)
	}
	if len(paper.Authors) != 2 {
		t.Errorf("expected 2 authors, got %d", len(paper.Authors))
	}
	if len(paper.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(paper.Categories))
	}
}

func TestExtractID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://arxiv.org/abs/2301.00001v1", "2301.00001v1"},
		{"http://arxiv.org/abs/cs/0001001", "cs/0001001"},
		{"invalid-id", "invalid-id"},
	}

	for _, tc := range tests {
		result := extractID(tc.input)
		if result != tc.expected {
			t.Errorf("extractID(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestExtractAuthors(t *testing.T) {
	authors := []atomAuthor{
		{Name: "John Doe"},
		{Name: "  Jane Smith  "},
		{Name: ""},
	}

	result := extractAuthors(authors)

	if len(result) != 2 {
		t.Errorf("expected 2 authors, got %d", len(result))
	}
	if result[0] != "John Doe" {
		t.Errorf("expected 'John Doe', got %q", result[0])
	}
	if result[1] != "Jane Smith" {
		t.Errorf("expected 'Jane Smith', got %q", result[1])
	}
}

func TestExtractCategories(t *testing.T) {
	categories := []atomCategory{
		{Term: "cs.AI"},
		{Term: "cs.LG"},
		{Term: ""},
	}

	result := extractCategories(categories)

	if len(result) != 2 {
		t.Errorf("expected 2 categories, got %d", len(result))
	}
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello world  ", "hello world"},
		{"hello\nworld", "hello world"},
		{"hello  world", "hello world"},
		{"  multi\n  line\n  text  ", "multi line text"},
	}

	for _, tc := range tests {
		result := cleanText(tc.input)
		if result != tc.expected {
			t.Errorf("cleanText(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
