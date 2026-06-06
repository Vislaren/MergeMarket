// Package store is the user-data service's data-access layer over PostgreSQL
// (DATABASE_SCHEMA.md). It owns per-user rows in wishlist_items, price_alerts,
// and purchases. Every operation is scoped to a user_id so one user can never
// read or mutate another's data. Access is behind the Repository interface so the
// HTTP handlers can be unit-tested with a fake; the pgx implementation lives in
// postgres.go.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a row does not exist (or is not owned by the
// requesting user). ErrAlreadyExists is returned on a duplicate wishlist entry.
// ErrUnknownProduct is returned when a referenced product_id does not exist.
var (
	ErrNotFound       = errors.New("store: not found")
	ErrAlreadyExists  = errors.New("store: already exists")
	ErrUnknownProduct = errors.New("store: unknown product")
)

// StorePrice is one store's price for a product (wishlist price comparison).
type StorePrice struct {
	Store     string  `json:"store"`
	Price     float64 `json:"price"`
	TotalCost float64 `json:"total_cost"`
}

// WishlistItem is one wishlist entry with the per-store price comparison.
type WishlistItem struct {
	WishlistID string       `json:"wishlist_id"`
	ProductID  string       `json:"product_id"`
	Title      string       `json:"title"`
	ImageURL   string       `json:"image_url"`
	Stores     []StorePrice `json:"stores"`
	AddedAt    time.Time    `json:"added_at"`
}

// WishlistAdded is the POST /wishlist response.
type WishlistAdded struct {
	WishlistID string    `json:"wishlist_id"`
	AddedAt    time.Time `json:"added_at"`
}

// Alert is one price alert.
type Alert struct {
	AlertID        string    `json:"alert_id"`
	ProductID      string    `json:"product_id"`
	Title          string    `json:"title"`
	ThresholdPrice float64   `json:"threshold_price"`
	Currency       string    `json:"currency"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// AlertCreated is the POST /alerts response.
type AlertCreated struct {
	AlertID   string    `json:"alert_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Transaction is one saving event on the savings dashboard.
type Transaction struct {
	ProductID string    `json:"product_id"`
	Title     string    `json:"title"`
	Saved     float64   `json:"saved"`
	BoughtAt  time.Time `json:"bought_at"`
}

// Savings is the GET /savings response.
type Savings struct {
	TotalSaved   float64       `json:"total_saved"`
	Currency     string        `json:"currency"`
	Transactions []Transaction `json:"transactions"`
}

// Repository is the user-data persistence surface. Every method takes the
// authenticated userID.
type Repository interface {
	// Wishlist.
	AddWishlist(ctx context.Context, userID, productID string) (WishlistAdded, error)
	ListWishlist(ctx context.Context, userID string) ([]WishlistItem, error)
	RemoveWishlist(ctx context.Context, userID, wishlistID string) error

	// Alerts.
	CreateAlert(ctx context.Context, userID, productID string, threshold float64, currency string) (AlertCreated, error)
	ListAlerts(ctx context.Context, userID string) ([]Alert, error)
	DeleteAlert(ctx context.Context, userID, alertID string) error

	// Savings.
	Savings(ctx context.Context, userID string) (Savings, error)

	// Close releases any underlying resources.
	Close() error
}
