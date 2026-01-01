package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/api"
	"github.com/1psychoQAQ/genesis-pipeline/internal/parser/arxiv"
	"github.com/1psychoQAQ/genesis-pipeline/internal/storage"
)

func main() {
	port := flag.String("port", "8080", "API server port")
	flag.Parse()

	log.Println("Genesis API Server starting...")

	// Connect to database
	ctx := context.Background()
	cfg := storage.DefaultConfig()

	pool, err := storage.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL")

	// Run migrations
	if err := storage.Migrate(ctx, pool); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Create dependencies
	repo := storage.NewPaperRepository(pool)
	client := arxiv.NewClient()
	handler := api.NewHandler(repo, client)

	// Setup routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Create server
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      logMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("API server listening on http://localhost:%s", *port)
	log.Println("Endpoints:")
	log.Println("  GET  /api/papers       - List papers")
	log.Println("  GET  /api/papers/:id   - Get paper by ID")
	log.Println("  GET  /api/papers/search?q= - Search papers")
	log.Println("  GET  /api/stats        - Pipeline statistics")
	log.Println("  POST /api/sync         - Trigger sync")
	log.Println("  GET  /health           - Health check")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
