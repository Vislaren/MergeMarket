// Package server wires the BFF's HTTP surface: a health probe, a minimal
// Prometheus metrics endpoint, the one aggregate product-detail view, and a
// catch-all reverse proxy that forwards every other /api/v1 request straight to
// the upstream API. It uses only the standard library.
package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/Vislaren/MergeMarket/services/bff/internal/config"
	"github.com/Vislaren/MergeMarket/services/bff/internal/upstream"
)

// Server holds the dependencies shared by all handlers.
type Server struct {
	cfg      config.Config
	logger   *slog.Logger
	upstream *upstream.Client
	proxy    http.Handler
	requests atomic.Int64
}

// New constructs a Server. The logger must be non-nil. It returns an error if
// the configured upstream URL cannot be parsed for the reverse proxy.
func New(cfg config.Config, logger *slog.Logger) (*Server, error) {
	target, err := url.Parse(cfg.UpstreamURL)
	if err != nil {
		return nil, fmt.Errorf("server: parse upstream url: %w", err)
	}
	return &Server{
		cfg:      cfg,
		logger:   logger,
		upstream: upstream.New(cfg.UpstreamURL),
		proxy:    httputil.NewSingleHostReverseProxy(target),
	}, nil
}

// Handler builds the fully-routed HTTP handler, wrapped with CORS, request
// counting, and logging middleware.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /metrics", s.handleMetrics)

	// The one shaped/aggregated view. More specific than the catch-all, so it
	// wins under Go 1.22 pattern precedence.
	mux.HandleFunc("GET /api/v1/products/{product_id}/detail", s.handleProductDetail)

	// Everything else is pure forwarding to the upstream API.
	mux.Handle("/", s.proxy)

	return s.withMiddleware(mux)
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

// withMiddleware adds permissive CORS, answers preflight, counts every request
// for /metrics, and logs method/path/status/duration.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.requests.Add(1)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// statusRecorder captures the response status code for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the status code before delegating.
func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// handleHealth serves GET /health.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": config.ServiceName,
		"version": config.Version,
	})
}

// handleMetrics serves GET /metrics in Prometheus text exposition format.
//
// The BFF is dependency-free, so this exposes a single hand-rolled counter
// rather than pulling in client_golang. A richer metric set can be added when
// the service is wired into Prometheus (A-12).
func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w,
		"# HELP bff_requests_total Total HTTP requests handled by the BFF.\n"+
			"# TYPE bff_requests_total counter\n"+
			"bff_requests_total %d\n",
		s.requests.Load(),
	)
}

// handleProductDetail serves GET /api/v1/products/{product_id}/detail — the
// aggregated history + truth-score + offers view.
func (s *Server) handleProductDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("product_id")
	detail, err := buildProductDetail(r.Context(), s.upstream, id)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "not_found", "product not found")
			return
		}
		s.logger.Error("aggregate product detail", "product_id", id, "error", err)
		writeError(w, http.StatusBadGateway, "upstream_error",
			"could not load product details")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes the canonical {error, message} body with the given status.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": code, "message": message})
}
