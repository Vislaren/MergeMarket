// Package fetcher downloads raw public proxy lists over HTTP and parses them
// into de-duplicated proxy addresses. Each source is independent: a single
// failing source never aborts the others (NFR-2, resilience).
package fetcher

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Vislaren/MergeMarket/services/proxy-validator/internal/proxy"
)

// HTTPDoer is the minimal HTTP surface fetcher needs, so tests can inject a
// stub instead of reaching the network.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Fetcher downloads and parses proxy lists from a fixed set of source URLs.
type Fetcher struct {
	sources []string
	client  HTTPDoer
	log     *slog.Logger
}

// New returns a Fetcher for the given sources. If client is nil a default
// HTTP client with a 15s timeout is used; if log is nil a no-op logger is used.
func New(sources []string, client HTTPDoer, log *slog.Logger) *Fetcher {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	if log == nil {
		log = slog.New(slog.NewTextHandler(noopWriter{}, nil))
	}
	return &Fetcher{sources: sources, client: client, log: log}
}

// FetchAll downloads every source, parses each into proxy addresses, and
// returns the merged, de-duplicated set. Errors from individual sources are
// logged and skipped; FetchAll only returns an error if every source failed.
func (f *Fetcher) FetchAll(ctx context.Context) ([]proxy.Addr, error) {
	var allLines []string
	var failures int

	for _, src := range f.sources {
		lines, err := f.fetchOne(ctx, src)
		if err != nil {
			f.log.WarnContext(ctx, "proxy source failed", "source", src, "error", err)
			failures++
			continue
		}
		f.log.InfoContext(ctx, "proxy source fetched", "source", src, "lines", len(lines))
		allLines = append(allLines, lines...)
	}

	if failures == len(f.sources) {
		return nil, fmt.Errorf("fetcher: all %d proxy sources failed", len(f.sources))
	}
	return proxy.ParseList(allLines), nil
}

// fetchOne downloads a single source and returns its raw, non-empty lines.
func (f *Fetcher) fetchOne(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetcher: build request: %w", err)
	}
	req.Header.Set("User-Agent", "MergeMarket-ProxyValidator/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetcher: get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetcher: %s returned status %d", url, resp.StatusCode)
	}

	var lines []string
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("fetcher: read %s: %w", url, err)
	}
	return lines, nil
}

// noopWriter discards log output for the default no-op logger.
type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }
