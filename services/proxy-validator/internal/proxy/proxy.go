// Package proxy defines the proxy address value type and the parsing rules
// used to turn raw public-proxy-list lines into validated "ip:port" entries.
package proxy

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Addr is a normalised proxy endpoint in canonical "ip:port" form.
type Addr struct {
	IP   string
	Port int
}

// String renders the address back into canonical "ip:port" form.
func (a Addr) String() string {
	return net.JoinHostPort(a.IP, strconv.Itoa(a.Port))
}

// Parse turns a single raw proxy-list line into an Addr. It tolerates the
// common noise found in free lists: surrounding whitespace, a leading
// "http://" / "https://" scheme, and trailing carriage returns. It rejects
// lines that are blank, commented (#), or not a valid host:port pair.
func Parse(line string) (Addr, error) {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimSuffix(s, "/")
	if s == "" || strings.HasPrefix(s, "#") {
		return Addr{}, fmt.Errorf("proxy: empty or comment line")
	}

	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return Addr{}, fmt.Errorf("proxy: %q is not host:port: %w", line, err)
	}
	if net.ParseIP(host) == nil {
		return Addr{}, fmt.Errorf("proxy: %q is not a valid IP", host)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return Addr{}, fmt.Errorf("proxy: %q is not a valid port", portStr)
	}
	return Addr{IP: host, Port: port}, nil
}

// ParseList parses many raw lines, silently dropping unparseable ones, and
// de-duplicates the result while preserving first-seen order. This is the
// shape callers want when ingesting a whole downloaded proxy list.
func ParseList(lines []string) []Addr {
	seen := make(map[string]struct{}, len(lines))
	out := make([]Addr, 0, len(lines))
	for _, line := range lines {
		a, err := Parse(line)
		if err != nil {
			continue
		}
		key := a.String()
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, a)
	}
	return out
}
