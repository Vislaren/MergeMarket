package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAllMergesAndDedups(t *testing.T) {
	srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("1.2.3.4:8080\n5.6.7.8:3128\n# comment\n"))
	}))
	defer srvA.Close()
	srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("5.6.7.8:3128\n9.9.9.9:80\ngarbage\n"))
	}))
	defer srvB.Close()

	f := New([]string{srvA.URL, srvB.URL}, srvA.Client(), nil)
	addrs, err := f.FetchAll(context.Background())
	if err != nil {
		t.Fatalf("FetchAll error = %v", err)
	}
	// 1.2.3.4, 5.6.7.8 (deduped across sources), 9.9.9.9 => 3 unique.
	if len(addrs) != 3 {
		t.Fatalf("got %d addrs, want 3: %#v", len(addrs), addrs)
	}
}

func TestFetchAllOneSourceFailsOthersSucceed(t *testing.T) {
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("1.2.3.4:8080\n"))
	}))
	defer good.Close()
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()

	f := New([]string{bad.URL, good.URL}, good.Client(), nil)
	addrs, err := f.FetchAll(context.Background())
	if err != nil {
		t.Fatalf("FetchAll should tolerate a single failing source, got %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("got %d addrs, want 1", len(addrs))
	}
}

func TestFetchAllAllSourcesFail(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer bad.Close()

	f := New([]string{bad.URL}, bad.Client(), nil)
	if _, err := f.FetchAll(context.Background()); err == nil {
		t.Fatal("FetchAll expected error when every source fails")
	}
}
