package middleware

import (
	"fmt"
	"net"
	"net/http"
)

func CheckCIDR(cidrStr string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			netIP, err := parseCIDRString(cidrStr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			clientIP, err := readClientIP(r)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if !isIPInCIDR(clientIP, netIP) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isIPInCIDR(clientIP net.IP, cidr *net.IPNet) bool {
	return cidr.Contains(clientIP)
}

func readClientIP(r *http.Request) (net.IP, error) {
	if ip := r.Header.Get("X-Real-IP"); ip == "" {
		return nil, fmt.Errorf("X-Real-IP header is missing")
	} else {
		parsedIP := net.ParseIP(ip)
		return parsedIP, nil
	}
}

func parseCIDRString(cidrStr string) (*net.IPNet, error) {
	_, ipnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return nil, err
	}

	return ipnet, nil
}
