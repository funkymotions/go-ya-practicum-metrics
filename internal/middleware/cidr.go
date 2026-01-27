package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/utils"
)

func CheckCIDR(cidrStr string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			netIP, err := utils.ParseCIDRString(cidrStr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			clientIP, err := readClientIP(r)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if !utils.IsIPInCIDR(clientIP, netIP) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func readClientIP(r *http.Request) (net.IP, error) {
	if ip := r.Header.Get("X-Real-IP"); ip == "" {
		return nil, fmt.Errorf("X-Real-IP header is missing")
	} else {
		parsedIP := net.ParseIP(ip)
		return parsedIP, nil
	}
}
