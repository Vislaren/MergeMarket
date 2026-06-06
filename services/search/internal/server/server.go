// Package server exposes the search-service HTTP surface: the API contract route
// GET /api/v1/search, plus GET /health. It validates query parameters and maps
// failures onto the canonical API_CONTRACTS.md error shapes.
package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Vislaren/MergeMarket/services/search/internal/search"
)

const serviceName = "search-service"

// Searcher performs a cached, scored product search.
type Searcher interface {
	Search(ctx context.Context, query, location string) (search.Response, error)
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// New builds an *http.Server with the search and health routes registered.
func New(addr, version string, svc Searcher) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(version))
	mux.HandleFunc("GET /api/v1/search", searchHandler(svc))

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func healthHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Service: serviceName, Version: version})
	}
}

func searchHandler(svc Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		query := strings.TrimSpace(req.URL.Query().Get("q"))
		if query == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{"missing_query", "q query parameter is required"})
			return
		}
		location := strings.TrimSpace(req.URL.Query().Get("location"))
		if location == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{"missing_query", "location query parameter is required"})
			return
		}

		res, err := svc.Search(req.Context(), query, location)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{"server_error", "could not run search"})
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

// Shutdown gracefully stops the server, bounded by a 5-second deadline.
func Shutdown(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
