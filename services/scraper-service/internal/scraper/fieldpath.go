package scraper

import (
	"strconv"
	"strings"
)

// lookup walks a dotted path through decoded JSON (map[string]any / []any).
// Numeric path segments index into arrays, e.g. "data.items.0.price". It
// returns the located value and whether the full path resolved.
func lookup(v any, path string) (any, bool) {
	if path == "" {
		return v, true
	}
	cur := v
	for _, seg := range strings.Split(path, ".") {
		switch node := cur.(type) {
		case map[string]any:
			next, ok := node[seg]
			if !ok {
				return nil, false
			}
			cur = next
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(node) {
				return nil, false
			}
			cur = node[idx]
		default:
			return nil, false
		}
	}
	return cur, true
}

// asString coerces a JSON scalar to a string. Numbers are rendered without a
// trailing ".0"; bools and nil yield "".
func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return ""
	}
}

// asFloat coerces a JSON scalar to a float64. Numeric strings (optionally with
// surrounding spaces, currency symbols, or thousands separators) are parsed.
func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case string:
		return parseNumeric(t)
	default:
		return 0, false
	}
}

// parseNumeric extracts a float from a noisy price string such as "$1,299.00".
func parseNumeric(s string) (float64, bool) {
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			b.WriteRune(r)
		}
	}
	cleaned := b.String()
	if cleaned == "" || cleaned == "-" || cleaned == "." {
		return 0, false
	}
	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}
