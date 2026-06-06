package server

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/Vislaren/MergeMarket/services/bff/internal/upstream"
)

// ProductDetail is the single shaped view the mobile Product Detail screen
// consumes: the price history, review truth score, and all store offers for one
// product, plus the cheapest offer and its deal score. This is the BFF's one
// piece of value-add — it replaces three client round-trips with one.
type ProductDetail struct {
	ProductID  string              `json:"product_id"`
	Title      string              `json:"title"`
	History    upstream.History    `json:"history"`
	TruthScore upstream.TruthScore `json:"truth_score"`
	Offers     []upstream.Offer    `json:"offers"`
	BestOffer  *upstream.Offer     `json:"best_offer"`
	DealScore  int                 `json:"deal_score"`
	StoreCount int                 `json:"store_count"`
}

// buildProductDetail aggregates the product-detail view for productID.
//
// History is fetched first because its title keys the offer search; a 404 there
// is the product's 404. The offer search and truth score are then fetched
// concurrently and are best-effort — a failure in either degrades to an empty
// section rather than failing the whole view, so the screen still renders the
// history. Offers are returned sorted by total cost ascending; the cheapest is
// the best offer and supplies the headline deal score.
func buildProductDetail(ctx context.Context, c *upstream.Client, productID string) (ProductDetail, error) {
	history, err := c.History(ctx, productID)
	if err != nil {
		return ProductDetail{}, err
	}

	var (
		wg     sync.WaitGroup
		search upstream.SearchResponse
		truth  upstream.TruthScore
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		// Best-effort: ignore the error, leave offers empty on failure.
		search, _ = c.Search(ctx, history.Title, "")
	}()
	go func() {
		defer wg.Done()
		truth, _ = c.TruthScore(ctx, productID)
	}()
	wg.Wait()

	offers := search.Results
	sort.SliceStable(offers, func(i, j int) bool {
		return offers[i].TotalCost < offers[j].TotalCost
	})

	detail := ProductDetail{
		ProductID:  productID,
		Title:      history.Title,
		History:    history,
		TruthScore: truth,
		Offers:     offers,
		StoreCount: distinctStores(offers),
	}
	if len(offers) > 0 {
		detail.BestOffer = &offers[0]
		detail.DealScore = offers[0].DealScore
	}
	return detail, nil
}

// distinctStores counts the unique store names across offers.
func distinctStores(offers []upstream.Offer) int {
	seen := make(map[string]struct{}, len(offers))
	for _, o := range offers {
		seen[o.Store] = struct{}{}
	}
	return len(seen)
}

// isNotFound reports whether err is an upstream 404.
func isNotFound(err error) bool {
	var nf *upstream.NotFoundError
	return errors.As(err, &nf)
}
