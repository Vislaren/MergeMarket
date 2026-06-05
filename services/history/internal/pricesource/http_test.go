package pricesource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/history/internal/store"
)

func TestExtractPrice(t *testing.T) {
	cases := []struct {
		name string
		body string
		want float64
		ok   bool
	}{
		{"json-ld", `{"name":"x","price": 129.99,"currency":"USD"}`, 129.99, true},
		{"json-ld string", `"price":"49.50"`, 49.50, true},
		{"og meta", `<meta property="product:price:amount" content="75.00">`, 75.00, true},
		{"itemprop", `<meta itemprop="price" content="10">`, 10, true},
		{"none", `<html>no price here</html>`, 0, false},
		{"zero ignored", `"price": 0`, 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := extractPrice([]byte(c.body))
			if ok != c.ok || (ok && got != c.want) {
				t.Errorf("extractPrice = (%v,%v), want (%v,%v)", got, ok, c.want, c.ok)
			}
		})
	}
}

func TestHTTPSource_CurrentPrice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><meta itemprop="price" content="59.99"></html>`))
	}))
	defer srv.Close()

	src := NewHTTP(5 * time.Second)
	r, err := src.CurrentPrice(context.Background(), store.Followed{URL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK || r.Price != 59.99 {
		t.Errorf("got %+v, want price 59.99", r)
	}
}

func TestHTTPSource_Non2xxNotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	src := NewHTTP(5 * time.Second)
	r, err := src.CurrentPrice(context.Background(), store.Followed{URL: srv.URL})
	if err != nil {
		t.Fatalf("non-2xx should not error: %v", err)
	}
	if r.OK {
		t.Errorf("expected OK=false for 404")
	}
}

func TestDBSource(t *testing.T) {
	r, _ := DBSource{}.CurrentPrice(context.Background(), store.Followed{LastPrice: 12.5, HasPrice: true})
	if !r.OK || r.Price != 12.5 {
		t.Errorf("got %+v", r)
	}
	r, _ = DBSource{}.CurrentPrice(context.Background(), store.Followed{HasPrice: false})
	if r.OK {
		t.Errorf("expected OK=false when no price")
	}
}
