// Package pricesource determines the current price of a followed product for the
// heartbeat. Two implementations are provided: DBSource reads the freshest price
// the scraper/normalization pipeline already persisted (reliable, the default),
// and HTTPSource re-fetches the product URL and best-effort extracts an embedded
// price ("scrape followed product URLs"). The Source interface keeps the runner
// decoupled and unit-testable.
package pricesource

import (
	"context"

	"github.com/Vislaren/MergeMarket/services/history/internal/store"
)

// Reading is the outcome of a price check. OK is false when no price could be
// determined (the heartbeat then skips that product rather than recording a
// bogus 0).
type Reading struct {
	Price    float64
	Shipping *float64
	OK       bool
}

// Source resolves the current price of a followed product.
type Source interface {
	CurrentPrice(ctx context.Context, f store.Followed) (Reading, error)
}

// DBSource returns the price already on the product row (products.last_price),
// kept current by the normalization-service. It performs no network I/O.
type DBSource struct{}

// CurrentPrice returns the product's last known price.
func (DBSource) CurrentPrice(_ context.Context, f store.Followed) (Reading, error) {
	if !f.HasPrice {
		return Reading{OK: false}, nil
	}
	shipping := f.LastShipping
	return Reading{Price: f.LastPrice, Shipping: &shipping, OK: true}, nil
}
