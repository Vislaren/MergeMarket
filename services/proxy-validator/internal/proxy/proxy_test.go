package proxy

import "testing"

func TestParseValid(t *testing.T) {
	cases := map[string]string{
		"plain":           "1.2.3.4:8080",
		"with scheme":     "http://1.2.3.4:8080",
		"https scheme":    "https://1.2.3.4:8080",
		"trailing slash":  "http://1.2.3.4:8080/",
		"surrounding ws":  "  1.2.3.4:8080  ",
		"carriage return": "1.2.3.4:8080\r",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			a, err := Parse(in)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", in, err)
			}
			if a.String() != "1.2.3.4:8080" {
				t.Errorf("Parse(%q) = %q, want 1.2.3.4:8080", in, a.String())
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	cases := []string{
		"", "   ", "# comment", "not-an-ip:80", "1.2.3.4", "1.2.3.4:99999",
		"1.2.3.4:abc", "999.999.999.999:80",
	}
	for _, in := range cases {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) expected error, got nil", in)
		}
	}
}

func TestParseListDedupAndDrop(t *testing.T) {
	lines := []string{
		"1.2.3.4:8080",
		"# a comment",
		"1.2.3.4:8080", // duplicate
		"garbage",
		"5.6.7.8:3128",
		"",
	}
	got := ParseList(lines)
	if len(got) != 2 {
		t.Fatalf("ParseList returned %d addrs, want 2: %#v", len(got), got)
	}
	if got[0].String() != "1.2.3.4:8080" || got[1].String() != "5.6.7.8:3128" {
		t.Errorf("ParseList preserved wrong order/values: %#v", got)
	}
}
