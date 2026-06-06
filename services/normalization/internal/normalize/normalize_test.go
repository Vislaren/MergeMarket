package normalize

import (
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/normalization/internal/queue"
)

func TestFromRaw_SkipsInvalidAndNormalizes(t *testing.T) {
	scrapedAt := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	raw := queue.RawResult{
		StoreID:   "jumia-cm",
		Store:     "Jumia",
		Query:     "phone",
		ScrapedAt: scrapedAt,
		Products: []queue.RawProduct{
			{ProductID: "1", Title: "  Cool   Phone ", Price: 19.999, Currency: "usd", Shipping: 2.5, URL: "https://x/y", ImageURL: "https://img"},
			{Title: "", Price: 10},                                       // no title → skipped
			{Title: "No price", Price: 0},                                // non-positive price → skipped
			{Title: "Neg ship", Price: 5, Shipping: -3, Currency: "eur"}, // shipping clamped to 0
			{Title: "Bad cur", Price: 8, Currency: "dollars"},            // bad currency → USD
		},
	}

	res := FromRaw(raw)

	if len(res.Products) != 3 {
		t.Fatalf("expected 3 valid products, got %d", len(res.Products))
	}

	p := res.Products[0]
	if p.Title != "Cool Phone" {
		t.Errorf("title not collapsed: %q", p.Title)
	}
	if p.Price != 20.00 {
		t.Errorf("price not rounded: %v", p.Price)
	}
	if p.Currency != "USD" {
		t.Errorf("currency not upper-cased: %q", p.Currency)
	}
	if p.TotalCost != 22.50 {
		t.Errorf("total_cost = %v, want 22.50", p.TotalCost)
	}
	if !p.ScrapedAt.Equal(scrapedAt) {
		t.Errorf("scraped_at not propagated")
	}

	if res.Products[1].Shipping != 0 {
		t.Errorf("negative shipping not clamped: %v", res.Products[1].Shipping)
	}
	if res.Products[1].Currency != "EUR" {
		t.Errorf("valid currency not preserved: %q", res.Products[1].Currency)
	}
	if res.Products[2].Currency != "USD" {
		t.Errorf("invalid currency not defaulted: %q", res.Products[2].Currency)
	}
}

func TestFromRaw_ZeroScrapedAtFilled(t *testing.T) {
	res := FromRaw(queue.RawResult{
		Store:    "S",
		Products: []queue.RawProduct{{Title: "t", Price: 1}},
	})
	if res.Products[0].ScrapedAt.IsZero() {
		t.Errorf("zero scraped_at should be filled with now")
	}
}

func TestRound2(t *testing.T) {
	cases := map[float64]float64{
		19.999: 20.00,
		19.994: 19.99,
		0.005:  0.01,
		2.5:    2.50,
	}
	for in, want := range cases {
		if got := round2(in); got != want {
			t.Errorf("round2(%v) = %v, want %v", in, got, want)
		}
	}
}
