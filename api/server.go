package api

import (
	"context"
	"encoding/json"
	"net/http"

	"gosearch/crawler"
	"gosearch/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Server holds shared dependencies for HTTP handlers.
type Server struct {
	Pool *pgxpool.Pool
}

// NewServer creates a Server with the given database pool.
func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{Pool: pool}
}

// SearchHandler handles GET /search?q=<query> and returns JSON results.
func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error": "missing query parameter 'q'"}`, http.StatusBadRequest)
		return
	}

	results, err := storage.SearchDocuments(context.Background(), s.Pool, query)
	if err != nil {
		http.Error(w, `{"error": "search failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// EnableCORS wraps a handler to allow requests from any origin (fine for development).
func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// CrawlRequest is the expected JSON body for POST /crawl
type CrawlRequest struct {
	SeedURL  string `json:"seed_url"`
	MaxPages int    `json:"max_pages"`
}

// CrawlHandler handles POST /crawl - starts a new crawl and saves results to the database.
func (s *Server) CrawlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "only POST method allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req CrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.SeedURL == "" {
		http.Error(w, `{"error": "seed_url is required"}`, http.StatusBadRequest)
		return
	}
	if req.MaxPages <= 0 {
		req.MaxPages = 5 // sensible default
	}

	c := crawler.NewCrawler(req.SeedURL, req.MaxPages)
	pages := c.Run()

	ctx := context.Background()
	saved := 0
	for _, page := range pages {
		if err := storage.SaveDocument(ctx, s.Pool, page.URL, page.Text); err == nil {
			saved++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"pages_crawled": len(pages),
		"pages_saved":   saved,
	})
}
