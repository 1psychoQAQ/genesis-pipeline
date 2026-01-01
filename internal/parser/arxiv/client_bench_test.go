package arxiv

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkFetchPapers(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClientWithOptions(server.Client(), server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.FetchPapers("test", 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExtractID(b *testing.B) {
	id := "http://arxiv.org/abs/2301.00001v1"
	for i := 0; i < b.N; i++ {
		extractID(id)
	}
}

func BenchmarkCleanText(b *testing.B) {
	text := "  This is a   multi-line\n\n  text with   lots of   spaces  "
	for i := 0; i < b.N; i++ {
		cleanText(text)
	}
}
