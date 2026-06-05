// Package server exposes the history-service's HTTP surface: the API contract
// route GET /api/v1/products/{product_id}/history, plus GET /health and the
// /stats observability endpoint. It uses the Go 1.22 method+path routing
// patterns so {product_id} is a first-class path value.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Vislaren/MergeMarket/services/history/internal/runner"
	"github.com/Vislaren/MergeMarket/services/history/internal/store"
)

const serviceName = "history-service"

// HistoryProvider serves the price-history query for a product.
type HistoryProvider interface {
	History(ctx context.Context, productID string) (store.HistoryResult, error)
}

// StatsProvider supplies the latest runner statistics.
type StatsProvider interface {
	Stats() runner.Stats
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

// New builds an *http.Server with the API, health, and stats routes registered.
func New(addr, version string, hist HistoryProvider, stats StatsProvider) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(version))
	mux.HandleFunc("GET /stats", statsHandler(stats))
	mux.HandleFunc("GET /api/v1/products/{product_id}/history", historyHandler(hist))

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

func statsHandler(stats StatsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, stats.Stats())
	}
}

func historyHandler(hist HistoryProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		productID := req.PathValue("product_id")
		if productID == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "product_id is required"})
			return
		}
		res, err := hist.History(req.Context(), productID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeJSON(w, http.StatusNotFound, errorResponse{"not_found", "product not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorResponse{"server_error", "could not load price history"})
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
