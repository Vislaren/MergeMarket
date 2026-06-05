package scraper

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/queue"
	"github.com/Vislaren/MergeMarket/services/scraper-service/internal/storeconfig"
)

func jsonAPIConfig(urlTmpl string) *storeconfig.StoreConfig {
	return &storeconfig.StoreConfig{
		StoreID:  "test",
		Name:     "Test Store",
		BaseURL:  "https://store.example",
		Currency: "USD",
		Mode:     storeconfig.ModeJSONAPI,
		Search:   storeconfig.Search{URLTemplate: urlTmpl},
		JSON: &storeconfig.JSONMapping{
			ResultsPath: "data.products",
			ProductID:   "id",
			Title:       "name",
			Price:       "price",
			Currency:    "currency",
			Shipping:    "shipping",
			ImageURL:    "image",
			URL:         "path",
		},
	}
}

func TestScrapeSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "phone" {
			t.Errorf("query q=%q, want phone", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"products":[
			{"id":"p1","name":"Phone X","price":199.99,"currency":"EUR","shipping":"$5.00","image":"http://cdn/x.jpg","path":"/p/p1"},
			{"id":"p2","name":"Phone Y","price":"1,299.00","image":"http://cdn/y.jpg","path":"/p/p2"},
			{"id":"p3","name":"","price":10}
		]}}`))
	}))
	defer srv.Close()

	eng := New(nil, 5*time.Second, nil)
	cfg := jsonAPIConfig(srv.URL + "?q={query}&loc={location}")
	res, err := eng.Scrape(context.Background(), cfg, queue.Job{StoreID: "test", Query: "phone", Location: "CM"})
	if err != nil {
		t.Fatalf("Scrape error: %v", err)
	}
	if len(res.Products) != 2 { // p3 dropped (empty title)
		t.Fatalf("got %d products, want 2: %+v", len(res.Products), res.Products)
	}
	p1 := res.Products[0]
	if p1.Price != 199.99 || p1.Currency != "EUR" || p1.Shipping != 5.00 {
		t.Errorf("p1 mapping wrong: %+v", p1)
	}
	if p1.URL != "https://store.example/p/p1" {
		t.Errorf("p1 URL = %q, want resolved absolute", p1.URL)
	}
	p2 := res.Products[1]
	if p2.Price != 1299.00 {
		t.Errorf("p2 price = %v, want 1299", p2.Price)
	}
	if p2.Currency != "USD" { // falls back to store currency
		t.Errorf("p2 currency = %q, want USD fallback", p2.Currency)
	}
	if res.Store != "Test Store" || res.Query != "phone" || res.ScrapedAt.IsZero() {
		t.Errorf("result envelope wrong: %+v", res)
	}
}

func TestScrapeBlocked(t *testing.T) {
	for _, code := range []int{http.StatusForbidden, http.StatusTooManyRequests} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(code)
		}))
		eng := New(nil, 5*time.Second, nil)
		cfg := jsonAPIConfig(srv.URL + "?q={query}")
		_, err := eng.Scrape(context.Background(), cfg, queue.Job{StoreID: "test", Query: "x"})
		if !errors.Is(err, ErrBlocked) {
			t.Errorf("status %d: err = %v, want ErrBlocked", code, err)
		}
		srv.Close()
	}
}

func TestScrapeNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	eng := New(nil, 5*time.Second, nil)
	cfg := jsonAPIConfig(srv.URL + "?q={query}")
	_, err := eng.Scrape(context.Background(), cfg, queue.Job{StoreID: "test", Query: "x"})
	if err == nil || errors.Is(err, ErrBlocked) {
		t.Errorf("500: err = %v, want a non-blocked error", err)
	}
}

func TestScrapeUnsupportedMode(t *testing.T) {
	eng := New(nil, time.Second, nil)
	cfg := &storeconfig.StoreConfig{StoreID: "h", Mode: storeconfig.ModeHTML, Search: storeconfig.Search{URLTemplate: "http://x"}}
	_, err := eng.Scrape(context.Background(), cfg, queue.Job{StoreID: "h"})
	if !errors.Is(err, ErrUnsupportedMode) {
		t.Errorf("err = %v, want ErrUnsupportedMode", err)
	}
}

func TestScrapeBadResultsPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"products":"not-an-array"}}`))
	}))
	defer srv.Close()
	eng := New(nil, 5*time.Second, nil)
	cfg := jsonAPIConfig(srv.URL + "?q={query}")
	if _, err := eng.Scrape(context.Background(), cfg, queue.Job{StoreID: "test", Query: "x"}); err == nil {
		t.Error("expected error when results_path is not an array")
	}
}

func TestRenderURLEscaping(t *testing.T) {
	got := renderURL("http://x/s?q={query}&c={location}", "red shoes & socks", "CM")
	want := "http://x/s?q=red+shoes+%26+socks&c=CM"
	if got != want {
		t.Errorf("renderURL = %q, want %q", got, want)
	}
}
