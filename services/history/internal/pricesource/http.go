package pricesource

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/Vislaren/MergeMarket/services/history/internal/store"
)

// userAgent identifies the heartbeat politely to retailers.
const userAgent = "MergeMarketBot/1.0 (+https://mergemarket.example/bot)"

// pricePatterns are tried in order to pull a price out of a product page. They
// cover JSON-LD / embedded JSON ("price": 12.34) and the common Open Graph /
// schema.org price meta tags. This is intentionally best-effort: a product page
// that exposes none of these yields OK=false and the product is skipped.
var pricePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)"price"\s*:\s*"?([0-9]+(?:\.[0-9]+)?)"?`),
	regexp.MustCompile(`(?i)property=["']product:price:amount["']\s+content=["']([0-9]+(?:\.[0-9]+)?)["']`),
	regexp.MustCompile(`(?i)itemprop=["']price["']\s+content=["']([0-9]+(?:\.[0-9]+)?)["']`),
}

// HTTPSource re-fetches a product URL and extracts an embedded price. The zero
// value is not usable; construct with NewHTTP.
type HTTPSource struct {
	client *http.Client
}

// NewHTTP builds an HTTPSource with the given per-request timeout.
func NewHTTP(timeout time.Duration) *HTTPSource {
	return &HTTPSource{client: &http.Client{Timeout: timeout}}
}

// CurrentPrice fetches the product URL and returns the first price it can
// extract. A non-2xx response or an unparseable body yields OK=false (not an
// error) so one unreadable page never aborts the heartbeat cycle (NFR-2). Only a
// transport-level failure is returned as an error.
func (h *HTTPSource) CurrentPrice(ctx context.Context, f store.Followed) (Reading, error) {
	if f.URL == "" {
		return Reading{OK: false}, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
	if err != nil {
		return Reading{OK: false}, fmt.Errorf("pricesource: build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := h.client.Do(req)
	if err != nil {
		return Reading{OK: false}, fmt.Errorf("pricesource: fetch %s: %w", f.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Reading{OK: false}, nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20)) // cap at 8 MiB
	if err != nil {
		return Reading{OK: false}, fmt.Errorf("pricesource: read body: %w", err)
	}

	price, ok := extractPrice(body)
	if !ok {
		return Reading{OK: false}, nil
	}
	return Reading{Price: price, OK: true}, nil
}

// extractPrice returns the first positive price matched by any pattern.
func extractPrice(body []byte) (float64, bool) {
	for _, re := range pricePatterns {
		m := re.FindSubmatch(body)
		if m == nil {
			continue
		}
		f, err := strconv.ParseFloat(string(m[1]), 64)
		if err != nil || f <= 0 {
			continue
		}
		return f, true
	}
	return 0, false
}
