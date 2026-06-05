// Package validator tests whether a proxy is actually usable by routing a
// real HTTP request through it to a known endpoint and checking the response.
package validator

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/proxy"
)

// ClientFactory builds an *http.Client that routes all traffic through the
// given proxy "ip:port". It is injectable so tests can supply a client that
// targets an httptest server instead of a live proxy.
type ClientFactory func(proxyAddr string, timeout time.Duration) (*http.Client, error)

// Validator checks proxies against a fixed test URL within a per-proxy timeout.
type Validator struct {
	testURL string
	timeout time.Duration
	factory ClientFactory
}

// New returns a Validator. If factory is nil, the default HTTP-proxy client
// factory is used (suitable for production).
func New(testURL string, timeout time.Duration, factory ClientFactory) *Validator {
	if factory == nil {
		factory = DefaultClientFactory
	}
	return &Validator{testURL: testURL, timeout: timeout, factory: factory}
}

// Validate returns nil if the proxy successfully relayed a request to the test
// URL and the endpoint replied with a non-error status (< 400). Otherwise it
// returns a named error describing the failure.
func (v *Validator) Validate(ctx context.Context, addr proxy.Addr) error {
	client, err := v.factory(addr.String(), v.timeout)
	if err != nil {
		return fmt.Errorf("validator: build client for %s: %w", addr, err)
	}
	defer client.CloseIdleConnections()

	reqCtx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, v.testURL, nil)
	if err != nil {
		return fmt.Errorf("validator: build request: %w", err)
	}
	req.Header.Set("User-Agent", "MergeMarket-ProxyValidator/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("validator: proxy %s failed: %w", addr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("validator: proxy %s got status %d", addr, resp.StatusCode)
	}
	return nil
}

// DefaultClientFactory builds a real HTTP client whose transport tunnels all
// requests through the given proxy address.
func DefaultClientFactory(proxyAddr string, timeout time.Duration) (*http.Client, error) {
	proxyURL, err := url.Parse("http://" + proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("validator: parse proxy url: %w", err)
	}
	transport := &http.Transport{
		Proxy:               http.ProxyURL(proxyURL),
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: timeout,
	}
	return &http.Client{Transport: transport, Timeout: timeout}, nil
}
