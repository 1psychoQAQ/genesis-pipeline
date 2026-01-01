package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
	"github.com/1psychoQAQ/genesis-pipeline/internal/parser"
	"github.com/1psychoQAQ/genesis-pipeline/internal/storage"
)

// Handler holds the API dependencies.
type Handler struct {
	repo     *storage.PaperRepository
	provider parser.Provider
}

// NewHandler creates a new API handler.
func NewHandler(repo *storage.PaperRepository, provider parser.Provider) *Handler {
	return &Handler{
		repo:     repo,
		provider: provider,
	}
}

// RegisterRoutes registers all API routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/papers", h.handlePapers)
	mux.HandleFunc("/api/papers/", h.handlePaperByID)
	mux.HandleFunc("/api/papers/search", h.handleSearch)
	mux.HandleFunc("/api/stats", h.handleStats)
	mux.HandleFunc("/api/sync", h.handleSync)
	mux.HandleFunc("/health", h.handleHealth)
}

// GET /api/papers - List papers with pagination
func (h *Handler) handlePapers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	papers, err := h.repo.List(ctx, limit, offset)
	if err != nil {
		log.Printf("Error listing papers: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Print paper links to console
	printPaperLinks(papers)

	respondJSON(w, http.StatusOK, map[string]any{
		"papers": papers,
		"limit":  limit,
		"offset": offset,
		"count":  len(papers),
	})
}

// GET /api/papers/:id - Get paper by ID
func (h *Handler) handlePaperByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path: /api/papers/2301.00001
	id := strings.TrimPrefix(r.URL.Path, "/api/papers/")
	if id == "" || id == "search" {
		http.Error(w, "Paper ID required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	paper, err := h.repo.GetByID(ctx, id)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Paper not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting paper: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, paper)
}

// GET /api/papers/search?q=query - Search papers
func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	papers, err := h.repo.Search(ctx, query, limit)
	if err != nil {
		log.Printf("Error searching papers: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Print paper links to console
	printPaperLinks(papers)

	respondJSON(w, http.StatusOK, map[string]any{
		"query":  query,
		"papers": papers,
		"count":  len(papers),
	})
}

// GET /api/stats - Get pipeline statistics
func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	count, err := h.repo.Count(ctx)
	if err != nil {
		log.Printf("Error getting count: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	latest, err := h.repo.GetLatestUpdateTime(ctx)
	if err != nil && err != storage.ErrNotFound {
		log.Printf("Error getting latest update: %v", err)
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"total_papers":  count,
		"last_sync":     latest,
		"database":      "PostgreSQL",
		"data_source":   "ArXiv API",
	})
}

// POST /api/sync - Trigger paper sync
func (h *Handler) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		query = "machine learning"
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	// Fetch papers from ArXiv
	papers, err := h.provider.FetchPapers(query, limit)
	if err != nil {
		log.Printf("Error fetching papers: %v", err)
		http.Error(w, "Failed to fetch papers", http.StatusInternalServerError)
		return
	}

	// Save to database
	if err := h.repo.SaveBatch(ctx, papers); err != nil {
		log.Printf("Error saving papers: %v", err)
		http.Error(w, "Failed to save papers", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"message": "Sync completed",
		"query":   query,
		"fetched": len(papers),
	})
}

// GET /health - Health check
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// PaperResponse is the JSON response for a paper.
type PaperResponse struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Abstract   string    `json:"abstract"`
	Authors    []string  `json:"authors"`
	Categories []string  `json:"categories"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ToPaperResponse converts a model.Paper to API response.
func ToPaperResponse(p model.Paper) PaperResponse {
	return PaperResponse{
		ID:         p.ID,
		Title:      p.Title,
		Abstract:   p.Abstract,
		Authors:    p.Authors,
		Categories: p.Categories,
		UpdatedAt:  p.UpdatedAt,
	}
}

// printPaperLinks prints paper links to console
func printPaperLinks(papers []model.Paper) {
	if len(papers) == 0 {
		return
	}
	log.Println("")
	log.Println("════════════════════════════════════════════════════════════")
	log.Printf("  📚 Found %d papers:", len(papers))
	log.Println("════════════════════════════════════════════════════════════")
	for i, p := range papers {
		log.Printf("  [%d] %s", i+1, p.Title)
		log.Printf("      📄 https://arxiv.org/abs/%s", p.ID)
		log.Printf("      📥 https://arxiv.org/pdf/%s.pdf", p.ID)
		log.Println("")
	}
	log.Println("════════════════════════════════════════════════════════════")
}
