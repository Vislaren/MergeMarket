package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/userdata/internal/store"
	"github.com/Vislaren/MergeMarket/services/userdata/internal/token"
)

// fakeRepo implements store.Repository for handler tests.
type fakeRepo struct {
	added     store.WishlistAdded
	addErr    error
	items     []store.WishlistItem
	listErr   error
	removeErr error
	created   store.AlertCreated
	createErr error
	alerts    []store.Alert
	delErr    error
	sav       store.Savings
	savErr    error

	lastUserID string
}

func (f *fakeRepo) AddWishlist(_ context.Context, userID, _ string) (store.WishlistAdded, error) {
	f.lastUserID = userID
	return f.added, f.addErr
}
func (f *fakeRepo) ListWishlist(_ context.Context, userID string) ([]store.WishlistItem, error) {
	f.lastUserID = userID
	return f.items, f.listErr
}
func (f *fakeRepo) RemoveWishlist(_ context.Context, _, _ string) error { return f.removeErr }
func (f *fakeRepo) CreateAlert(_ context.Context, _, _ string, _ float64, _ string) (store.AlertCreated, error) {
	return f.created, f.createErr
}
func (f *fakeRepo) ListAlerts(_ context.Context, _ string) ([]store.Alert, error) {
	return f.alerts, f.listErr
}
func (f *fakeRepo) DeleteAlert(_ context.Context, _, _ string) error { return f.delErr }
func (f *fakeRepo) Savings(_ context.Context, _ string) (store.Savings, error) {
	return f.sav, f.savErr
}
func (f *fakeRepo) Close() error { return nil }

// fakeVerifier accepts the token "good" as user "u1" and rejects everything else.
type fakeVerifier struct{ err error }

func (v fakeVerifier) Verify(tok string) (token.Claims, error) {
	if v.err != nil {
		return token.Claims{}, v.err
	}
	if tok == "good" {
		return token.Claims{UserID: "u1", Email: "u1@x.com"}, nil
	}
	return token.Claims{}, token.ErrInvalidToken
}

func req(method, target, body, auth string) *http.Request {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	if auth != "" {
		r.Header.Set("Authorization", "Bearer "+auth)
	}
	return r
}

func serve(repo store.Repository, v Verifier, r *http.Request) *httptest.ResponseRecorder {
	srv := New(":0", "test", repo, v)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, r)
	return rec
}

func TestRequiresAuth(t *testing.T) {
	rec := serve(&fakeRepo{}, fakeVerifier{}, req(http.MethodGet, "/api/v1/wishlist", "", ""))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", rec.Code)
	}
}

func TestRejectsBadToken(t *testing.T) {
	rec := serve(&fakeRepo{}, fakeVerifier{}, req(http.MethodGet, "/api/v1/wishlist", "", "bad"))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", rec.Code)
	}
}

func TestExpiredTokenMapsToTokenExpired(t *testing.T) {
	rec := serve(&fakeRepo{}, fakeVerifier{err: token.ErrExpired}, req(http.MethodGet, "/api/v1/wishlist", "", "good"))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", rec.Code)
	}
	var e errorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &e)
	if e.Error != "token_expired" {
		t.Errorf("error = %q, want token_expired", e.Error)
	}
}

func TestListWishlistEmptyReturnsArray(t *testing.T) {
	rec := serve(&fakeRepo{items: nil}, fakeVerifier{}, req(http.MethodGet, "/api/v1/wishlist", "", "good"))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Errorf("empty wishlist should serialize items as [], got %s", rec.Body.String())
	}
}

func TestAddWishlistThreadsUserID(t *testing.T) {
	repo := &fakeRepo{added: store.WishlistAdded{WishlistID: "w1", AddedAt: time.Now()}}
	rec := serve(repo, fakeVerifier{}, req(http.MethodPost, "/api/v1/wishlist", `{"product_id":"p1"}`, "good"))
	if rec.Code != http.StatusCreated {
		t.Fatalf("code = %d, want 201", rec.Code)
	}
	if repo.lastUserID != "u1" {
		t.Errorf("handler did not thread authenticated user id, got %q", repo.lastUserID)
	}
}

func TestAddWishlistMissingProduct(t *testing.T) {
	rec := serve(&fakeRepo{}, fakeVerifier{}, req(http.MethodPost, "/api/v1/wishlist", `{}`, "good"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestAddWishlistConflict(t *testing.T) {
	repo := &fakeRepo{addErr: store.ErrAlreadyExists}
	rec := serve(repo, fakeVerifier{}, req(http.MethodPost, "/api/v1/wishlist", `{"product_id":"p1"}`, "good"))
	if rec.Code != http.StatusConflict {
		t.Fatalf("code = %d, want 409", rec.Code)
	}
}

func TestAddWishlistUnknownProduct(t *testing.T) {
	repo := &fakeRepo{addErr: store.ErrUnknownProduct}
	rec := serve(repo, fakeVerifier{}, req(http.MethodPost, "/api/v1/wishlist", `{"product_id":"nope"}`, "good"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestRemoveWishlistNotFound(t *testing.T) {
	repo := &fakeRepo{removeErr: store.ErrNotFound}
	rec := serve(repo, fakeVerifier{}, req(http.MethodDelete, "/api/v1/wishlist/w1", "", "good"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", rec.Code)
	}
}

func TestRemoveWishlistSuccess(t *testing.T) {
	rec := serve(&fakeRepo{}, fakeVerifier{}, req(http.MethodDelete, "/api/v1/wishlist/w1", "", "good"))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
}

func TestCreateAlertValidation(t *testing.T) {
	// threshold_price must be > 0.
	rec := serve(&fakeRepo{}, fakeVerifier{}, req(http.MethodPost, "/api/v1/alerts", `{"product_id":"p1","threshold_price":0}`, "good"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestCreateAlertSuccess(t *testing.T) {
	repo := &fakeRepo{created: store.AlertCreated{AlertID: "a1", CreatedAt: time.Now()}}
	rec := serve(repo, fakeVerifier{}, req(http.MethodPost, "/api/v1/alerts", `{"product_id":"p1","threshold_price":99.5,"currency":"USD"}`, "good"))
	if rec.Code != http.StatusCreated {
		t.Fatalf("code = %d, want 201", rec.Code)
	}
}

func TestDeleteAlertNotFound(t *testing.T) {
	repo := &fakeRepo{delErr: store.ErrNotFound}
	rec := serve(repo, fakeVerifier{}, req(http.MethodDelete, "/api/v1/alerts/a1", "", "good"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", rec.Code)
	}
}

func TestSavingsSuccess(t *testing.T) {
	repo := &fakeRepo{sav: store.Savings{TotalSaved: 42.5, Currency: "USD", Transactions: []store.Transaction{}}}
	rec := serve(repo, fakeVerifier{}, req(http.MethodGet, "/api/v1/savings", "", "good"))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	var s store.Savings
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.TotalSaved != 42.5 || s.Currency != "USD" {
		t.Errorf("unexpected savings: %+v", s)
	}
}

func TestHealthNoAuth(t *testing.T) {
	rec := serve(&fakeRepo{}, fakeVerifier{err: errors.New("should not be called")}, req(http.MethodGet, "/health", "", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
}
