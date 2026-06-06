// Package fixtures defines the response data types for every API contract in
// project_docs/api/API_CONTRACTS.md and the static sample data the mock server
// returns. Field names and JSON tags mirror the contracts exactly so the
// Flutter app and BFF can be developed against shapes identical to the real
// services.
//
// All money values use minor-unit-free numbers in the response currency
// (XAF for the "CM" sample locale). total_cost always equals price + shipping.
package fixtures

import "time"

// iso8601 formats a time as an RFC 3339 / ISO 8601 string in UTC.
func iso8601(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

// HealthResponse is the GET /health body.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

// ErrorResponse is the canonical error shape for all endpoints.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Err builds an ErrorResponse.
func Err(code, message string) ErrorResponse {
	return ErrorResponse{Error: code, Message: message}
}

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

// AuthRequest is the body for register/login.
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest is the body for the refresh endpoint.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse is returned by register, login, and refresh.
type TokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

// Token returns a sample token bundle that expires one hour from now.
func Token() TokenResponse {
	return TokenResponse{
		Token:        "mock.jwt.access-token",
		RefreshToken: "mock.jwt.refresh-token",
		ExpiresAt:    iso8601(time.Now().Add(time.Hour)),
	}
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// SearchResult is a single store offer in a search response.
type SearchResult struct {
	ProductID    string  `json:"product_id"`
	Title        string  `json:"title"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	Shipping     float64 `json:"shipping"`
	TotalCost    float64 `json:"total_cost"`
	ImageURL     string  `json:"image_url"`
	Store        string  `json:"store"`
	AffiliateURL string  `json:"affiliate_url"`
	DealScore    int     `json:"deal_score"`
	ScrapedAt    string  `json:"scraped_at"`
}

// SearchResponse is the GET /api/v1/search body.
type SearchResponse struct {
	Query     string         `json:"query"`
	Results   []SearchResult `json:"results"`
	Cached    bool           `json:"cached"`
	LatencyMs int            `json:"latency_ms"`
}

// offer is a small helper to keep total_cost = price + shipping consistent.
func offer(id, title string, price, shipping float64, store, img, aff string, deal int) SearchResult {
	return SearchResult{
		ProductID:    id,
		Title:        title,
		Price:        price,
		Currency:     "XAF",
		Shipping:     shipping,
		TotalCost:    price + shipping,
		ImageURL:     img,
		Store:        store,
		AffiliateURL: aff,
		DealScore:    deal,
		ScrapedAt:    iso8601(time.Now().Add(-3 * time.Minute)),
	}
}

// Search returns a sample multi-store result set for the given query.
func Search(query string) SearchResponse {
	return SearchResponse{
		Query: query,
		Results: []SearchResult{
			offer("prod-001", "Samsung Galaxy A54 128GB", 245000, 5000,
				"Jumia", "https://img.example/galaxy-a54.jpg",
				"https://aff.example/jumia/prod-001", 88),
			offer("prod-001", "Samsung Galaxy A54 128GB", 239900, 8000,
				"Kilimall", "https://img.example/galaxy-a54.jpg",
				"https://aff.example/kilimall/prod-001", 81),
			offer("prod-001", "Samsung Galaxy A54 (128GB)", 252000, 0,
				"AfricShop", "https://img.example/galaxy-a54.jpg",
				"https://aff.example/africshop/prod-001", 74),
		},
		Cached:    true,
		LatencyMs: 42,
	}
}

// ---------------------------------------------------------------------------
// Price history & truth score
// ---------------------------------------------------------------------------

// PricePoint is one historical price observation.
type PricePoint struct {
	Price      float64 `json:"price"`
	Currency   string  `json:"currency"`
	RecordedAt string  `json:"recorded_at"`
}

// HistoryResponse is the GET /api/v1/products/{id}/history body.
type HistoryResponse struct {
	ProductID string       `json:"product_id"`
	Title     string       `json:"title"`
	History   []PricePoint `json:"history"`
	Average6m float64      `json:"average_6m"`
	Lowest30d float64      `json:"lowest_30d"`
}

// History returns six monthly price points (oldest first) for a product,
// along with the 6-month average and the lowest price in the last 30 days.
func History(productID string) HistoryResponse {
	prices := []float64{268000, 261000, 255000, 258000, 249000, 245000}
	now := time.Now()

	points := make([]PricePoint, 0, len(prices))
	var sum, lowest30d float64
	lowest30d = prices[len(prices)-1]
	for i, p := range prices {
		monthsAgo := len(prices) - 1 - i
		points = append(points, PricePoint{
			Price:      p,
			Currency:   "XAF",
			RecordedAt: iso8601(now.AddDate(0, -monthsAgo, 0)),
		})
		sum += p
	}

	return HistoryResponse{
		ProductID: productID,
		Title:     "Samsung Galaxy A54 128GB",
		History:   points,
		Average6m: round2(sum / float64(len(prices))),
		Lowest30d: lowest30d,
	}
}

// TruthScoreResponse is the GET /api/v1/products/{id}/truth-score body.
type TruthScoreResponse struct {
	ProductID      string `json:"product_id"`
	Score          int    `json:"score"`
	Sentiment      string `json:"sentiment"`
	FakeReviewRisk string `json:"fake_review_risk"`
	Summary        string `json:"summary"`
}

// TruthScore returns a sample review-authenticity score for a product.
func TruthScore(productID string) TruthScoreResponse {
	return TruthScoreResponse{
		ProductID:      productID,
		Score:          82,
		Sentiment:      "positive",
		FakeReviewRisk: "low",
		Summary:        "Reviews are consistent across stores with low duplication; buyers praise battery life and value.",
	}
}

// ---------------------------------------------------------------------------
// Wishlist
// ---------------------------------------------------------------------------

// WishlistStore is one store's offer for a wishlisted product.
type WishlistStore struct {
	Store     string  `json:"store"`
	Price     float64 `json:"price"`
	TotalCost float64 `json:"total_cost"`
}

// WishlistItem is one entry in the wishlist.
type WishlistItem struct {
	WishlistID string          `json:"wishlist_id"`
	ProductID  string          `json:"product_id"`
	Title      string          `json:"title"`
	ImageURL   string          `json:"image_url"`
	Stores     []WishlistStore `json:"stores"`
	AddedAt    string          `json:"added_at"`
}

// WishlistResponse is the GET /api/v1/wishlist body.
type WishlistResponse struct {
	Items []WishlistItem `json:"items"`
}

// CreatedWishlist is the 201 body for POST /api/v1/wishlist.
type CreatedWishlist struct {
	WishlistID string `json:"wishlist_id"`
	AddedAt    string `json:"added_at"`
}

// Wishlist returns a sample wishlist with multi-store tracking per product.
func Wishlist() WishlistResponse {
	now := time.Now()
	return WishlistResponse{
		Items: []WishlistItem{
			{
				WishlistID: "wl-001",
				ProductID:  "prod-001",
				Title:      "Samsung Galaxy A54 128GB",
				ImageURL:   "https://img.example/galaxy-a54.jpg",
				Stores: []WishlistStore{
					{Store: "Jumia", Price: 245000, TotalCost: 250000},
					{Store: "Kilimall", Price: 239900, TotalCost: 247900},
				},
				AddedAt: iso8601(now.AddDate(0, 0, -5)),
			},
			{
				WishlistID: "wl-002",
				ProductID:  "prod-014",
				Title:      "Anker PowerCore 20000mAh",
				ImageURL:   "https://img.example/anker-20k.jpg",
				Stores: []WishlistStore{
					{Store: "Jumia", Price: 28500, TotalCost: 31500},
				},
				AddedAt: iso8601(now.AddDate(0, 0, -12)),
			},
		},
	}
}

// CreatedWishlistEntry returns the 201 body for a newly added wishlist item.
func CreatedWishlistEntry() CreatedWishlist {
	return CreatedWishlist{
		WishlistID: "wl-003",
		AddedAt:    iso8601(time.Now()),
	}
}

// ---------------------------------------------------------------------------
// Alerts
// ---------------------------------------------------------------------------

// Alert is one price-threshold alert.
type Alert struct {
	AlertID        string  `json:"alert_id"`
	ProductID      string  `json:"product_id"`
	Title          string  `json:"title"`
	ThresholdPrice float64 `json:"threshold_price"`
	Currency       string  `json:"currency"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
}

// AlertsResponse is the GET /api/v1/alerts body.
type AlertsResponse struct {
	Alerts []Alert `json:"alerts"`
}

// CreatedAlert is the 201 body for POST /api/v1/alerts.
type CreatedAlert struct {
	AlertID   string `json:"alert_id"`
	CreatedAt string `json:"created_at"`
}

// Alerts returns a sample set of active price alerts.
func Alerts() AlertsResponse {
	now := time.Now()
	return AlertsResponse{
		Alerts: []Alert{
			{
				AlertID:        "al-001",
				ProductID:      "prod-001",
				Title:          "Samsung Galaxy A54 128GB",
				ThresholdPrice: 230000,
				Currency:       "XAF",
				IsActive:       true,
				CreatedAt:      iso8601(now.AddDate(0, 0, -4)),
			},
			{
				AlertID:        "al-002",
				ProductID:      "prod-014",
				Title:          "Anker PowerCore 20000mAh",
				ThresholdPrice: 25000,
				Currency:       "XAF",
				IsActive:       false,
				CreatedAt:      iso8601(now.AddDate(0, 0, -9)),
			},
		},
	}
}

// CreatedAlertEntry returns the 201 body for a newly created alert.
func CreatedAlertEntry() CreatedAlert {
	return CreatedAlert{
		AlertID:   "al-003",
		CreatedAt: iso8601(time.Now()),
	}
}

// ---------------------------------------------------------------------------
// Savings dashboard
// ---------------------------------------------------------------------------

// SavingsTransaction is one realised saving from a purchase.
type SavingsTransaction struct {
	ProductID string  `json:"product_id"`
	Title     string  `json:"title"`
	Saved     float64 `json:"saved"`
	BoughtAt  string  `json:"bought_at"`
}

// SavingsResponse is the GET /api/v1/savings body.
type SavingsResponse struct {
	TotalSaved   float64              `json:"total_saved"`
	Currency     string               `json:"currency"`
	Transactions []SavingsTransaction `json:"transactions"`
}

// Savings returns a sample gamified savings summary.
func Savings() SavingsResponse {
	now := time.Now()
	txns := []SavingsTransaction{
		{ProductID: "prod-001", Title: "Samsung Galaxy A54 128GB", Saved: 23000, BoughtAt: iso8601(now.AddDate(0, 0, -20))},
		{ProductID: "prod-009", Title: "JBL Tune 510BT Headphones", Saved: 6500, BoughtAt: iso8601(now.AddDate(0, 0, -40))},
		{ProductID: "prod-014", Title: "Anker PowerCore 20000mAh", Saved: 4000, BoughtAt: iso8601(now.AddDate(0, 0, -55))},
	}
	var total float64
	for _, t := range txns {
		total += t.Saved
	}
	return SavingsResponse{
		TotalSaved:   total,
		Currency:     "XAF",
		Transactions: txns,
	}
}

// round2 rounds a float to two decimal places.
func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
