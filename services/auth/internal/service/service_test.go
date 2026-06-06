package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/auth/internal/session"
	"github.com/Vislaren/MergeMarket/services/auth/internal/store"
	"github.com/Vislaren/MergeMarket/services/auth/internal/token"
)

type fakeRepo struct {
	user store.User
	err  error
}

func (f *fakeRepo) CreateUser(context.Context, string, string) (store.User, error) {
	return f.user, f.err
}
func (f *fakeRepo) UserByEmail(context.Context, string) (store.User, error) { return f.user, f.err }

type fakeSessions struct {
	data session.Data
	err  error
}

func (f *fakeSessions) Create(context.Context, session.Data) (string, error) { return "refresh", f.err }
func (f *fakeSessions) Consume(context.Context, string) (session.Data, error) {
	return f.data, f.err
}

func TestRegisterMapsDuplicateEmail(t *testing.T) {
	auth := New(&fakeRepo{err: store.ErrEmailExists}, &fakeSessions{}, token.NewIssuer([]byte("12345678901234567890123456789012"), time.Hour), time.Hour, slog.Default())
	_, err := auth.Register(context.Background(), "a@example.com", "password123")
	if !errors.Is(err, ErrEmailExists) {
		t.Fatalf("Register() error = %v, want ErrEmailExists", err)
	}
}

func TestLoginRejectsBadPassword(t *testing.T) {
	hash := "$2a$10$o3p.W2g.PCuzofvxFe9txuOOGbmOPpSpw9V8m.2p2nbF6lGtxKc1C"
	auth := New(&fakeRepo{user: store.User{ID: "u1", Email: "a@example.com", PasswordHash: hash}}, &fakeSessions{}, token.NewIssuer([]byte("12345678901234567890123456789012"), time.Hour), time.Hour, slog.Default())
	_, err := auth.Login(context.Background(), "a@example.com", "wrongpass")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}
