// Package server exposes the user-data service HTTP surface: the wishlist, alert,
// and savings routes from API_CONTRACTS.md, plus GET /health. Every API route is
// JWT-protected; the authenticated user_id is extracted from the verified token
// and threaded into the store so users only ever touch their own data.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Vislaren/MergeMarket/services/userdata/internal/store"
	"github.com/Vislaren/MergeMarket/services/userdata/internal/token"
)

const serviceName = "userdata-service"

// Verifier validates a bearer token and returns its claims.
type Verifier interface {
	Verify(tokenString string) (token.Claims, error)
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

type wishlistResponse struct {
	Items []store.WishlistItem `json:"items"`
}

type alertsResponse struct {
	Alerts []store.Alert `json:"alerts"`
}

type addWishlistRequest struct {
	ProductID string `json:"product_id"`
}

type createAlertRequest struct {
	ProductID      string  `json:"product_id"`
	ThresholdPrice float64 `json:"threshold_price"`
	Currency       string  `json:"currency"`
}

// New builds an *http.Server with all routes registered.
func New(addr, version string, repo store.Repository, v Verifier) *http.Server {
	h := &handlers{repo: repo, verifier: v}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(version))

	mux.HandleFunc("GET /api/v1/wishlist", h.authed(h.listWishlist))
	mux.HandleFunc("POST /api/v1/wishlist", h.authed(h.addWishlist))
	mux.HandleFunc("DELETE /api/v1/wishlist/{wishlist_id}", h.authed(h.removeWishlist))

	mux.HandleFunc("GET /api/v1/alerts", h.authed(h.listAlerts))
	mux.HandleFunc("POST /api/v1/alerts", h.authed(h.createAlert))
	mux.HandleFunc("DELETE /api/v1/alerts/{alert_id}", h.authed(h.deleteAlert))

	mux.HandleFunc("GET /api/v1/savings", h.authed(h.savings))

	return &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
}

type handlers struct {
	repo     store.Repository
	verifier Verifier
}

// authed wraps an authenticated handler: it verifies the bearer token and passes
// the resolved user_id through. Any verification failure short-circuits with 401.
func (h *handlers) authed(fn func(w http.ResponseWriter, r *http.Request, userID string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := bearerToken(r)
		if raw == "" {
			writeJSON(w, http.StatusUnauthorized, errorResponse{"unauthorized", "missing bearer token"})
			return
		}
		claims, err := h.verifier.Verify(raw)
		if err != nil {
			if errors.Is(err, token.ErrExpired) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{"token_expired", "access token has expired"})
				return
			}
			writeJSON(w, http.StatusUnauthorized, errorResponse{"unauthorized", "invalid access token"})
			return
		}
		fn(w, r, claims.UserID)
	}
}

func (h *handlers) listWishlist(w http.ResponseWriter, r *http.Request, userID string) {
	items, err := h.repo.ListWishlist(r.Context(), userID)
	if err != nil {
		writeServerError(w, "could not load wishlist")
		return
	}
	if items == nil {
		items = []store.WishlistItem{}
	}
	writeJSON(w, http.StatusOK, wishlistResponse{Items: items})
}

func (h *handlers) addWishlist(w http.ResponseWriter, r *http.Request, userID string) {
	var req addWishlistRequest
	if !decode(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.ProductID) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "product_id is required"})
		return
	}
	added, err := h.repo.AddWishlist(r.Context(), userID, req.ProductID)
	switch {
	case errors.Is(err, store.ErrAlreadyExists):
		writeJSON(w, http.StatusConflict, errorResponse{"already_in_wishlist", "product is already in the wishlist"})
	case errors.Is(err, store.ErrUnknownProduct):
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "unknown product_id"})
	case err != nil:
		writeServerError(w, "could not add to wishlist")
	default:
		writeJSON(w, http.StatusCreated, added)
	}
}

func (h *handlers) removeWishlist(w http.ResponseWriter, r *http.Request, userID string) {
	id := r.PathValue("wishlist_id")
	err := h.repo.RemoveWishlist(r.Context(), userID, id)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{"not_found", "wishlist item not found"})
	case err != nil:
		writeServerError(w, "could not remove wishlist item")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *handlers) listAlerts(w http.ResponseWriter, r *http.Request, userID string) {
	alerts, err := h.repo.ListAlerts(r.Context(), userID)
	if err != nil {
		writeServerError(w, "could not load alerts")
		return
	}
	if alerts == nil {
		alerts = []store.Alert{}
	}
	writeJSON(w, http.StatusOK, alertsResponse{Alerts: alerts})
}

func (h *handlers) createAlert(w http.ResponseWriter, r *http.Request, userID string) {
	var req createAlertRequest
	if !decode(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.ProductID) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "product_id is required"})
		return
	}
	if req.ThresholdPrice <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "threshold_price must be greater than 0"})
		return
	}
	created, err := h.repo.CreateAlert(r.Context(), userID, req.ProductID, req.ThresholdPrice, req.Currency)
	switch {
	case errors.Is(err, store.ErrUnknownProduct):
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "unknown product_id"})
	case err != nil:
		writeServerError(w, "could not create alert")
	default:
		writeJSON(w, http.StatusCreated, created)
	}
}

func (h *handlers) deleteAlert(w http.ResponseWriter, r *http.Request, userID string) {
	id := r.PathValue("alert_id")
	err := h.repo.DeleteAlert(r.Context(), userID, id)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{"not_found", "alert not found"})
	case err != nil:
		writeServerError(w, "could not delete alert")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *handlers) savings(w http.ResponseWriter, r *http.Request, userID string) {
	s, err := h.repo.Savings(r.Context(), userID)
	if err != nil {
		writeServerError(w, "could not load savings")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func healthHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Service: serviceName, Version: version})
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return ""
}

// decode reads a JSON body into dst, writing a 400 and returning false on
// failure.
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "request body must be valid JSON"})
		return false
	}
	return true
}

func writeServerError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, errorResponse{"server_error", msg})
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
