// Package search is the search-service orchestrator. It ties the Redis result
// cache to the PostgreSQL product repository with a stale-while-revalidate
// strategy (ARCHITECTURE.md §10): a fresh cache hit is served immediately; a
// stale hit is served immediately while a background refresh runs; a miss reads
// the database, scores the results (Deal Meter), and populates the cache.
package search

import (
	"context"
	"log/slog"
	"time"

	"github.com/Vislaren/MergeMarket/services/search/internal/cache"
	"github.com/Vislaren/MergeMarket/services/search/internal/store"
)

// Response is the full search payload returned to the client
// (API_CONTRACTS.md GET /api/v1/search).
type Response struct {
	Query     string          `json:"query"`
	Results   []store.Product `json:"results"`
	Cached    bool            `json:"cached"`
	LatencyMs int64           `json:"latency_ms"`
}

// Config carries the cache tuning the orchestrator needs.
type Config struct {
	CachePrefix     string
	CacheTTL        time.Duration
	CacheStaleAfter time.Duration
	MaxResults      int
}

// Service orchestrates cached, scored product searches.
type Service struct {
	repo  store.Repository
	cache cache.Cache
	cfg   Config
	log   *slog.Logger
	now   func() time.Time
}

// New builds a search Service.
func New(repo store.Repository, c cache.Cache, cfg Config, log *slog.Logger) *Service {
	return &Service{repo: repo, cache: c, cfg: cfg, log: log, now: time.Now}
}

// Search returns scored results for the query, using the cache when possible.
// The returned Response carries the measured server-side latency and whether the
// results were served from cache.
func (s *Service) Search(ctx context.Context, query, location string) (Response, error) {
	start := s.now()
	key := cache.Key(s.cfg.CachePrefix, query, location)

	if entry, ok, err := s.cache.Get(ctx, key); err != nil {
		s.log.Warn("cache get failed; falling back to database", "error", err)
	} else if ok {
		if s.now().Sub(entry.CachedAt) > s.cfg.CacheStaleAfter {
			s.revalidate(query, location, key) // stale-while-revalidate
		}
		return s.respond(query, entry.Results, true, start), nil
	}

	results, err := s.fetchAndCache(ctx, query, location, key)
	if err != nil {
		return Response{}, err
	}
	return s.respond(query, results, false, start), nil
}

// fetchAndCache reads from the database, scores the results, and writes the
// cache. A cache-write failure is logged but not fatal — the results are still
// returned.
func (s *Service) fetchAndCache(ctx context.Context, query, location, key string) ([]store.Product, error) {
	results, err := s.repo.Search(ctx, query, location, s.cfg.MaxResults)
	if err != nil {
		return nil, err
	}
	scoreDeals(results)
	if err := s.cache.Set(ctx, key, cache.Entry{Results: results, CachedAt: s.now()}, s.cfg.CacheTTL); err != nil {
		s.log.Warn("cache set failed", "error", err)
	}
	return results, nil
}

// revalidate refreshes a stale cache entry in the background. It uses a fresh,
// bounded context (not the request context, which is canceled once the response
// is written).
func (s *Service) revalidate(query, location, key string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := s.fetchAndCache(ctx, query, location, key); err != nil {
			s.log.Warn("background revalidation failed", "error", err)
		}
	}()
}

func (s *Service) respond(query string, results []store.Product, cached bool, start time.Time) Response {
	if results == nil {
		results = []store.Product{}
	}
	return Response{
		Query:     query,
		Results:   results,
		Cached:    cached,
		LatencyMs: s.now().Sub(start).Milliseconds(),
	}
}

// scoreDeals assigns each result a 0–100 Deal Meter score based on its total
// cost relative to the cheapest and most expensive result in the same set: the
// cheapest scores 100, the most expensive scores 0, the rest scale linearly.
//
// This is a deterministic heuristic placeholder. A richer Deal Meter (comparing
// against the product's own price history and review authenticity) belongs to a
// future task (the truth-score service, A-15); keeping it here means the
// contract's required deal_score field is always populated meaningfully.
func scoreDeals(results []store.Product) {
	if len(results) == 0 {
		return
	}
	min, max := results[0].TotalCost, results[0].TotalCost
	for _, r := range results {
		if r.TotalCost < min {
			min = r.TotalCost
		}
		if r.TotalCost > max {
			max = r.TotalCost
		}
	}
	span := max - min
	for i := range results {
		if span <= 0 {
			results[i].DealScore = 100 // only one price point — it is the best deal
			continue
		}
		results[i].DealScore = int((max-results[i].TotalCost)/span*100 + 0.5)
	}
}
