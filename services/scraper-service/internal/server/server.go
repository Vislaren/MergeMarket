// Package server exposes the scraper-service's HTTP surface. Per the MergeMarket
// API contract every service exposes GET /health; this service additionally
// surfaces worker-pool counters under /stats for observability.
package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/worker"
)

const serviceName = "scraper-service"

// StatsProvider supplies the latest worker-pool statistics.
type StatsProvider interface {
	Stats() worker.Stats
}

// healthResponse matches the API_CONTRACTS.md /health shape.
type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// New builds an *http.Server listening on addr ("host:port" or ":port") with the
// /health and /stats routes registered.
func New(addr, version string, stats StatsProvider) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(version))
	mux.HandleFunc("/stats", statsHandler(stats))

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func healthHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status:  "ok",
			Service: serviceName,
			Version: version,
		})
	}
}

func statsHandler(stats StatsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, stats.Stats())
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
