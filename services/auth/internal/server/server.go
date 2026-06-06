// Package server exposes the auth-service HTTP API and health endpoint.
package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Vislaren/MergeMarket/services/auth/internal/service"
	"github.com/Vislaren/MergeMarket/services/auth/internal/token"
)

const serviceName = "auth-service"

// Authenticator is the HTTP-facing auth application interface.
type Authenticator interface {
	Register(ctx context.Context, email, password string) (token.Pair, error)
	Login(ctx context.Context, email, password string) (token.Pair, error)
	Refresh(ctx context.Context, refreshToken string) (token.Pair, error)
}

type credentialRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
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

// New builds the TLS-only auth-service HTTP server.
func New(addr, version string, auth Authenticator) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(version))
	mux.HandleFunc("POST /api/v1/auth/register", registerHandler(auth))
	mux.HandleFunc("POST /api/v1/auth/login", loginHandler(auth))
	mux.HandleFunc("POST /api/v1/auth/refresh", refreshHandler(auth))

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}
}

func healthHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Service: serviceName, Version: version})
	}
}

func registerHandler(auth Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var body credentialRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "request body must be valid JSON"})
			return
		}
		pair, err := auth.Register(req.Context(), body.Email, body.Password)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidInput):
				writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", err.Error()})
			case errors.Is(err, service.ErrEmailExists):
				writeJSON(w, http.StatusConflict, errorResponse{"email_exists", "email is already registered"})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{"server_error", "could not register user"})
			}
			return
		}
		writeJSON(w, http.StatusCreated, pair)
	}
}

func loginHandler(auth Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var body credentialRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "request body must be valid JSON"})
			return
		}
		pair, err := auth.Login(req.Context(), body.Email, body.Password)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{"invalid_credentials", "email or password is incorrect"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorResponse{"server_error", "could not log in"})
			return
		}
		writeJSON(w, http.StatusOK, pair)
	}
}

func refreshHandler(auth Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var body refreshRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_input", "request body must be valid JSON"})
			return
		}
		pair, err := auth.Refresh(req.Context(), body.RefreshToken)
		if err != nil {
			if errors.Is(err, service.ErrTokenExpired) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{"token_expired", "refresh token is expired or invalid"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorResponse{"server_error", "could not refresh token"})
			return
		}
		writeJSON(w, http.StatusOK, pair)
	}
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

// Shutdown gracefully stops the server within five seconds.
func Shutdown(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
