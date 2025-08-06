package middleware

import (
	"net/http"
	"strings"
)

type contentTypeMiddleware struct{}

func NewContentTypeMiddleware() *contentTypeMiddleware {
	return &contentTypeMiddleware{}
}

func (m *contentTypeMiddleware) CheckContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
