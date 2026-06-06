// Package affiliate implements MergeMarket's Affiliate Link Injection module
// (FR-6). It wraps an outbound product URL with retailer-specific affiliate
// parameters so that clicks earn commission. Two mechanisms are supported per
// store and combine: a redirect/deep-link template (a URL containing a {url}
// placeholder) and/or a set of query parameters appended to the link.
//
// The mapping is data-driven: a JSON config file keyed by the scraper's store id
// configures each retailer, with an optional set of default parameters applied
// when a store has no explicit entry. This keeps adding an affiliate program a
// config change, not a code change.
package affiliate

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
)

// placeholderEnc is replaced with the URL-encoded product URL; placeholderRaw
// with the un-encoded product URL. Templates use these to build deep links such
// as "https://go.skimresources.com/?id=123&url={url}".
const (
	placeholderEnc = "{url}"
	placeholderRaw = "{url_raw}"
)

// storeRule configures affiliate injection for one store.
type storeRule struct {
	// Template, if set, is the deep-link template with {url} / {url_raw}
	// placeholders. When empty the product URL itself is the base.
	Template string `json:"template"`
	// Params are query parameters appended to the resulting link (e.g. a partner
	// tag). They override default params for this store.
	Params map[string]string `json:"params"`
}

// fileFormat is the on-disk affiliate config shape.
type fileFormat struct {
	// DefaultParams are appended for stores without an explicit entry (or whose
	// entry sets no Params of its own).
	DefaultParams map[string]string `json:"default_params"`
	// Stores maps a scraper store id to its rule.
	Stores map[string]storeRule `json:"stores"`
}

// Injector applies affiliate rules to product URLs. The zero value is a no-op
// injector (returns URLs unchanged), which is a safe default when no config is
// supplied.
type Injector struct {
	defaultParams map[string]string
	stores        map[string]storeRule
}

// New returns an empty (no-op) Injector.
func New() *Injector {
	return &Injector{stores: map[string]storeRule{}}
}

// Load builds an Injector from a JSON config file. An empty path returns a no-op
// Injector with no error (injection simply leaves URLs unchanged). A missing
// file at a non-empty path is a configuration error.
func Load(path string) (*Injector, error) {
	if strings.TrimSpace(path) == "" {
		return New(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("affiliate: read config %s: %w", path, err)
	}
	var ff fileFormat
	if err := json.Unmarshal(data, &ff); err != nil {
		return nil, fmt.Errorf("affiliate: parse config %s: %w", path, err)
	}
	if ff.Stores == nil {
		ff.Stores = map[string]storeRule{}
	}
	return &Injector{defaultParams: ff.DefaultParams, stores: ff.Stores}, nil
}

// Inject returns the affiliate-wrapped URL for a product from storeID. An empty
// rawURL yields an empty string. When no rule matches and there are no default
// params, the URL is returned unchanged.
func (in *Injector) Inject(storeID, rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	rule, hasRule := in.stores[storeID]

	// A configured store fully specifies its own behaviour (template and/or
	// params); default_params is only the catch-all for un-configured stores.
	if hasRule {
		base := rawURL
		if rule.Template != "" {
			base = renderTemplate(rule.Template, rawURL)
		}
		return appendParams(base, rule.Params)
	}

	return appendParams(rawURL, in.defaultParams)
}

// renderTemplate substitutes the encoded/raw URL placeholders into a deep-link
// template.
func renderTemplate(tmpl, rawURL string) string {
	r := strings.NewReplacer(
		placeholderEnc, url.QueryEscape(rawURL),
		placeholderRaw, rawURL,
	)
	return r.Replace(tmpl)
}

// appendParams adds params to a URL's query string. On a parse failure it
// returns the base unchanged so a malformed link never blocks the pipeline.
// Parameters are applied in sorted key order for deterministic output.
func appendParams(base string, params map[string]string) string {
	if len(params) == 0 {
		return base
	}
	u, err := url.Parse(base)
	if err != nil {
		return base
	}
	q := u.Query()
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		q.Set(k, params[k])
	}
	u.RawQuery = q.Encode()
	return u.String()
}
