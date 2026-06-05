// Package server wires the HTTP routes that mirror every endpoint in
// project_docs/api/API_CONTRACTS.md and serves them with the static fixtures
// from the fixtures package. It uses only the standard library: Go 1.22 method
// + wildcard routing patterns, encoding/json, and log/slog.
//
// Handlers are intentionally simple. A few deterministic "sentinel" inputs let
// clients exercise non-200 paths without any server state (see the per-handler
// GoDoc). Everything else returns a successful sample response.
package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Vislaren/MergeMarket/services/mock-server/internal/config"
	"github.com/Vislaren/MergeMarket/services/mock-server/internal/fixtures"
)

// Server holds the dependencies shared by all handlers.
type Server struct {
	cfg    config.Config
	logger *slog.Logger
}

// New constructs a Server. The logger must be non-nil.
func New(cfg config.Config, logger *slog.Logger) *Server {
	return &Server{cfg: cfg, logger: logger}
}

// Handler builds the fully-routed HTTP handler, wrapped with CORS and request
// logging middleware.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", s.handleHealth)

	// Auth
	mux.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
	mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.handleRefresh)

	// Search
	mux.HandleFunc("GET /api/v1/search", s.handleSearch)

	// Products
	mux.HandleFunc("GET /api/v1/products/{product_id}/history", s.handleHistory)
	mux.HandleFunc("GET /api/v1/products/{product_id}/truth-score", s.handleTruthScore)

	// Wishlist
	mux.HandleFunc("GET /api/v1/wishlist", s.handleWishlistList)
	mux.HandleFunc("POST /api/v1/wishlist", s.handleWishlistAdd)
	mux.HandleFunc("DELETE /api/v1/wishlist/{wishlist_id}", s.handleWishlistDelete)

	// Alerts
	mux.HandleFunc("GET /api/v1/alerts", s.handleAlertsList)
	mux.HandleFunc("POST /api/v1/alerts", s.handleAlertCreate)
	mux.HandleFunc("DELETE /api/v1/alerts/{alert_id}", s.handleAlertDelete)

	// Savings
	mux.HandleFunc("GET /api/v1/savings", s.handleSavings)

	return s.withMiddleware(mux)
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

// withMiddleware adds permissive CORS (so a Flutter web/dev client can call the
// mock from any origin), answers CORS preflight requests, and logs every
// request with method, path, status, and duration.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, fixtures.HealthResponse{
		Status:  "ok",
		Service: config.ServiceName,
		Version: config.Version,
	})
}

// handleRegister serves POST /api/v1/auth/register.
//
// Sentinels: missing email/password → 400 invalid_input;
// email "taken@mergemarket.app" → 409 email_exists; otherwise 201.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req fixtures.AuthRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, fixtures.Err("invalid_input", "email and password are required"))
		return
	}
	if req.Email == "taken@mergemarket.app" {
		writeJSON(w, http.StatusConflict, fixtures.Err("email_exists", "an account with this email already exists"))
		return
	}
	writeJSON(w, http.StatusCreated, fixtures.Token())
}

// handleLogin serves POST /api/v1/auth/login.
//
// Sentinels: password "wrongpassword" (or missing credentials) → 401
// invalid_credentials; otherwise 200.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req fixtures.AuthRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Email == "" || req.Password == "" || req.Password == "wrongpassword" {
		writeJSON(w, http.StatusUnauthorized, fixtures.Err("invalid_credentials", "email or password is incorrect"))
		return
	}
	writeJSON(w, http.StatusOK, fixtures.Token())
}

// handleRefresh serves POST /api/v1/auth/refresh.
//
// Sentinels: refresh_token "expired" (or missing) → 401 token_expired;
// otherwise 200.
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req fixtures.RefreshRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.RefreshToken == "" || req.RefreshToken == "expired" {
		writeJSON(w, http.StatusUnauthorized, fixtures.Err("token_expired", "refresh token is invalid or expired"))
		return
	}
	writeJSON(w, http.StatusOK, fixtures.Token())
}

// handleSearch serves GET /api/v1/search.
//
// Sentinels: missing q → 400 missing_query; q "timeout" → 504 timeout;
// otherwise 200 with sample multi-store results.
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusBadRequest, fixtures.Err("missing_query", "query parameter 'q' is required"))
		return
	}
	if q == "timeout" {
		writeJSON(w, http.StatusGatewayTimeout, fixtures.Err("timeout", "Results took too long. Try again."))
		return
	}
	writeJSON(w, http.StatusOK, fixtures.Search(q))
}

// handleHistory serves GET /api/v1/products/{product_id}/history.
//
// Sentinel: product_id "unknown" → 404 not_found; otherwise 200.
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("product_id")
	if id == "unknown" {
		writeJSON(w, http.StatusNotFound, fixtures.Err("not_found", "product not found"))
		return
	}
	writeJSON(w, http.StatusOK, fixtures.History(id))
}

// handleTruthScore serves GET /api/v1/products/{product_id}/truth-score.
//
// Sentinel: product_id "unknown" → 404 not_found; otherwise 200.
func (s *Server) handleTruthScore(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("product_id")
	if id == "unknown" {
		writeJSON(w, http.StatusNotFound, fixtures.Err("not_found", "product not found"))
		return
	}
	writeJSON(w, http.StatusOK, fixtures.TruthScore(id))
}

// handleWishlistList serves GET /api/v1/wishlist.
func (s *Server) handleWishlistList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, fixtures.Wishlist())
}

// handleWishlistAdd serves POST /api/v1/wishlist.
//
// Sentinels: missing product_id → 400 invalid_input;
// product_id "prod-001" → 409 already_in_wishlist; otherwise 201.
func (s *Server) handleWishlistAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID string `json:"product_id"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ProductID == "" {
		writeJSON(w, http.StatusBadRequest, fixtures.Err("invalid_input", "product_id is required"))
		return
	}
	if req.ProductID == "prod-001" {
		writeJSON(w, http.StatusConflict, fixtures.Err("already_in_wishlist", "product is already in the wishlist"))
		return
	}
	writeJSON(w, http.StatusCreated, fixtures.CreatedWishlistEntry())
}

// handleWishlistDelete serves DELETE /api/v1/wishlist/{wishlist_id}.
//
// Sentinel: wishlist_id "unknown" → 404 not_found; otherwise 204 (no body).
func (s *Server) handleWishlistDelete(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("wishlist_id") == "unknown" {
		writeJSON(w, http.StatusNotFound, fixtures.Err("not_found", "wishlist item not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleAlertsList serves GET /api/v1/alerts.
func (s *Server) handleAlertsList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, fixtures.Alerts())
}

// handleAlertCreate serves POST /api/v1/alerts.
//
// Sentinel: missing product_id or non-positive threshold_price → 400
// invalid_input; otherwise 201.
func (s *Server) handleAlertCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID      string  `json:"product_id"`
		ThresholdPrice float64 `json:"threshold_price"`
		Currency       string  `json:"currency"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ProductID == "" || req.ThresholdPrice <= 0 {
		writeJSON(w, http.StatusBadRequest, fixtures.Err("invalid_input", "product_id and a positive threshold_price are required"))
		return
	}
	writeJSON(w, http.StatusCreated, fixtures.CreatedAlertEntry())
}

// handleAlertDelete serves DELETE /api/v1/alerts/{alert_id}.
//
// Sentinel: alert_id "unknown" → 404 not_found; otherwise 204 (no body).
func (s *Server) handleAlertDelete(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("alert_id") == "unknown" {
		writeJSON(w, http.StatusNotFound, fixtures.Err("not_found", "alert not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleSavings serves GET /api/v1/savings.
func (s *Server) handleSavings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, fixtures.Savings())
}

// ---------------------------------------------------------------------------
// JSON helpers
// ---------------------------------------------------------------------------

// writeJSON encodes v as JSON with the given status code. If encoding fails it
// falls back to a 500 with the canonical error shape.
func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server_error","message":"failed to encode response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// decodeJSON decodes the request body into dst. On a malformed body it writes a
// 400 invalid_input error and returns false. An empty body decodes to the zero
// value and is treated as success so per-field validation can run.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Body == nil || r.ContentLength == 0 {
		return true
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, fixtures.Err("invalid_input", "request body is not valid JSON"))
		return false
	}
	return true
}
