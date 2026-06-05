package validator

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/proxy"
)

// roundTripFunc adapts a function into an http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newResp(code int) *http.Response {
	return &http.Response{StatusCode: code, Body: http.NoBody, Header: make(http.Header)}
}

func TestValidateGoodProxy(t *testing.T) {
	// Factory routes through a transport that always answers 204.
	factory := func(_ string, _ time.Duration) (*http.Client, error) {
		return &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return newResp(http.StatusNoContent), nil
		})}, nil
	}
	v := New("http://test/endpoint", time.Second, factory)
	if err := v.Validate(context.Background(), proxy.Addr{IP: "1.2.3.4", Port: 8080}); err != nil {
		t.Fatalf("Validate good proxy error = %v", err)
	}
}

func TestValidateBadStatus(t *testing.T) {
	factory := func(_ string, _ time.Duration) (*http.Client, error) {
		return &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return newResp(http.StatusForbidden), nil
		})}, nil
	}
	v := New("http://test/endpoint", time.Second, factory)
	err := v.Validate(context.Background(), proxy.Addr{IP: "1.2.3.4", Port: 8080})
	if err == nil || !strings.Contains(err.Error(), "status 403") {
		t.Fatalf("Validate should reject 403, got %v", err)
	}
}

func TestValidateTransportError(t *testing.T) {
	factory := func(_ string, _ time.Duration) (*http.Client, error) {
		return &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		})}, nil
	}
	v := New("http://test/endpoint", time.Second, factory)
	if err := v.Validate(context.Background(), proxy.Addr{IP: "1.2.3.4", Port: 8080}); err == nil {
		t.Fatal("Validate should return error when the proxy transport fails")
	}
}

func TestDefaultClientFactoryBuildsProxyClient(t *testing.T) {
	c, err := DefaultClientFactory("1.2.3.4:8080", time.Second)
	if err != nil {
		t.Fatalf("DefaultClientFactory error = %v", err)
	}
	tr, ok := c.Transport.(*http.Transport)
	if !ok || tr.Proxy == nil {
		t.Fatal("DefaultClientFactory must set an *http.Transport with a Proxy")
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	u, err := tr.Proxy(req)
	if err != nil || u == nil || u.Host != "1.2.3.4:8080" {
		t.Fatalf("transport proxy = %v (err %v), want host 1.2.3.4:8080", u, err)
	}
}
