package scraper

import (
	"encoding/json"
	"testing"
)

func decode(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return v
}

func TestLookup(t *testing.T) {
	root := decode(t, `{"data":{"products":[{"name":"A","price":{"value":10}}]}}`)
	tests := []struct {
		path   string
		want   any
		wantOK bool
	}{
		{"data.products.0.name", "A", true},
		{"data.products.0.price.value", float64(10), true},
		{"data.products.1.name", nil, false}, // index out of range
		{"data.missing", nil, false},
		{"", root, true},
	}
	for _, tt := range tests {
		got, ok := lookup(root, tt.path)
		if ok != tt.wantOK {
			t.Errorf("lookup(%q) ok=%v, want %v", tt.path, ok, tt.wantOK)
			continue
		}
		if tt.wantOK && tt.path != "" && got != tt.want {
			t.Errorf("lookup(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestAsString(t *testing.T) {
	tests := []struct {
		in   any
		want string
	}{
		{"hello", "hello"},
		{float64(1299), "1299"},
		{float64(12.5), "12.5"},
		{true, "true"},
		{nil, ""},
	}
	for _, tt := range tests {
		if got := asString(tt.in); got != tt.want {
			t.Errorf("asString(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestAsFloat(t *testing.T) {
	tests := []struct {
		in     any
		want   float64
		wantOK bool
	}{
		{float64(10.5), 10.5, true},
		{"$1,299.00", 1299.00, true},
		{"  42 ", 42, true},
		{"XAF 5000", 5000, true},
		{"free", 0, false},
		{"", 0, false},
		{true, 0, false},
	}
	for _, tt := range tests {
		got, ok := asFloat(tt.in)
		if ok != tt.wantOK {
			t.Errorf("asFloat(%v) ok=%v, want %v", tt.in, ok, tt.wantOK)
			continue
		}
		if tt.wantOK && got != tt.want {
			t.Errorf("asFloat(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
