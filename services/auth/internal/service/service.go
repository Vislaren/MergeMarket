// Package service contains the auth-service application logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/Vislaren/MergeMarket/services/auth/internal/session"
	"github.com/Vislaren/MergeMarket/services/auth/internal/store"
	"github.com/Vislaren/MergeMarket/services/auth/internal/token"
)

// ErrInvalidInput indicates malformed user input.
var ErrInvalidInput = errors.New("auth: invalid input")

// ErrEmailExists indicates that a registration email is already taken.
var ErrEmailExists = errors.New("auth: email exists")

// ErrInvalidCredentials indicates an invalid login attempt.
var ErrInvalidCredentials = errors.New("auth: invalid credentials")

// ErrTokenExpired indicates a missing, expired, or consumed refresh token.
var ErrTokenExpired = errors.New("auth: token expired")

// Auth coordinates users, sessions, and token issuance.
type Auth struct {
	users    store.Repository
	sessions session.Store
	issuer   interface {
		Issue(userID, email string) (string, time.Time, error)
	}
	refreshTokenTTL time.Duration
	log             *slog.Logger
}

// New creates an auth application service.
func New(users store.Repository, sessions session.Store, issuer *token.Issuer, refreshTokenTTL time.Duration, log *slog.Logger) *Auth {
	return &Auth{users: users, sessions: sessions, issuer: issuer, refreshTokenTTL: refreshTokenTTL, log: log}
}

// Register creates a user and returns access + refresh tokens.
func (a *Auth) Register(ctx context.Context, email, password string) (token.Pair, error) {
	if err := validateCredentials(email, password); err != nil {
		return token.Pair{}, err
	}
	user, err := a.users.CreateUser(ctx, email, password)
	if err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			return token.Pair{}, ErrEmailExists
		}
		return token.Pair{}, err
	}
	return a.issuePair(ctx, user.ID, user.Email)
}

// Login verifies credentials and returns access + refresh tokens.
func (a *Auth) Login(ctx context.Context, email, password string) (token.Pair, error) {
	if err := validateCredentials(email, password); err != nil {
		return token.Pair{}, ErrInvalidCredentials
	}
	user, err := a.users.UserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return token.Pair{}, ErrInvalidCredentials
		}
		return token.Pair{}, err
	}
	if !store.CheckPassword(user.PasswordHash, password) {
		return token.Pair{}, ErrInvalidCredentials
	}
	return a.issuePair(ctx, user.ID, user.Email)
}

// Refresh consumes a refresh token and returns a rotated token pair.
func (a *Auth) Refresh(ctx context.Context, refreshToken string) (token.Pair, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return token.Pair{}, ErrTokenExpired
	}
	data, err := a.sessions.Consume(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			return token.Pair{}, ErrTokenExpired
		}
		return token.Pair{}, err
	}
	return a.issuePair(ctx, data.UserID, data.Email)
}

func (a *Auth) issuePair(ctx context.Context, userID, email string) (token.Pair, error) {
	access, expires, err := a.issuer.Issue(userID, email)
	if err != nil {
		return token.Pair{}, err
	}
	refresh, err := a.sessions.Create(ctx, session.Data{UserID: userID, Email: email})
	if err != nil {
		return token.Pair{}, err
	}
	a.log.Info("auth token pair issued", "user_id", userID, "access_expires_at", expires)
	return token.Pair{Token: access, RefreshToken: refresh, ExpiresAt: expires}, nil
}

func validateCredentials(email, password string) error {
	if _, err := mail.ParseAddress(strings.TrimSpace(email)); err != nil {
		return fmt.Errorf("%w: invalid email", ErrInvalidInput)
	}
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidInput)
	}
	return nil
}
