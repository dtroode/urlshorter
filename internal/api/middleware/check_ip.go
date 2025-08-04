package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

type subnet struct {
	subnet *net.IPNet
}

func parseCIDR(cidr string) (*subnet, error) {
	if cidr == "" {
		return &subnet{}, nil
	}

	_, s, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subnet: %w", err)
	}

	return &subnet{
		subnet: s,
	}, nil
}

// Contains checks if the given IP is in the subnet.
// It determines whether the specified IP address falls within the configured subnet range.
//
// Parameters:
//   - ip: The IP address to check
//
// Returns true if the IP is in the subnet, false otherwise.
func (s *subnet) Contains(ip net.IP) bool {
	if s.subnet == nil {
		return false
	}

	return s.subnet.Contains(ip)
}

// CheckIP is a middleware that checks if the request's IP is in a trusted subnet.
// It provides IP-based access control for HTTP requests by validating the client's IP address.
type CheckIP struct {
	subnet *subnet
}

// NewCheckIP creates a new CheckIP middleware.
// It initializes IP-based access control with the specified CIDR subnet.
//
// Parameters:
//   - cidr: The CIDR notation subnet (e.g., "192.168.1.0/24")
//
// Returns a pointer to the newly created CheckIP middleware or an error if CIDR parsing fails.
func NewCheckIP(cidr string) (*CheckIP, error) {
	subnet, err := parseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	return &CheckIP{
		subnet: subnet,
	}, nil
}

// Handle is a middleware that checks if the request's IP is in a trusted subnet.
// It extracts the client IP from the request headers and validates it against the configured subnet.
// If the IP is not in the trusted subnet, it returns a 403 Forbidden response.
//
// Parameters:
//   - h: The next HTTP handler in the middleware chain
//
// Returns an HTTP handler that performs IP validation before calling the next handler.
func (m *CheckIP) Handle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := resolveIP(r, resolveIPOpts{UseHeader: true})
		if err != nil {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		if !m.subnet.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		h.ServeHTTP(w, r)
	})
}

type resolveIPOpts struct {
	UseHeader bool
}

func resolveIP(r *http.Request, opts resolveIPOpts) (net.IP, error) {
	if !opts.UseHeader {
		addr := r.RemoteAddr

		ipStr, _, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, fmt.Errorf("ip is nil")
		}

		return ip, nil
	} else {
		ipStr := r.Header.Get("X-Real-IP")

		ip := net.ParseIP(ipStr)
		if ip == nil {
			ips := r.Header.Get("X-Forwarded-For")

			ipStrs := strings.Split(ips, ",")
			ipStr = ipStrs[0]

			ip = net.ParseIP(ipStr)
		}
		if ip == nil {
			return nil, fmt.Errorf("ip is nil")
		}

		return ip, nil
	}
}
