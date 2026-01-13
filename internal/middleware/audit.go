package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type AuditMiddleware struct {
	audit  auditPublisher
	logger *zap.Logger
}

type auditPublisher interface {
	Notify(metrics interface{}, ipAddress string)
}

type auditResponseWriter struct {
	StatusCode int
	http.ResponseWriter
	Size int
	Body []byte
}

func (arw *auditResponseWriter) WriteHeader(code int) {
	arw.StatusCode = code
	arw.ResponseWriter.WriteHeader(code)
}

func (arw *auditResponseWriter) Write(b []byte) (int, error) {
	arw.Body = b
	size, err := arw.ResponseWriter.Write(b)
	arw.Size += size
	return size, err
}

func NewAuditMiddleware(logger *zap.Logger, audit auditPublisher) *AuditMiddleware {
	return &AuditMiddleware{
		audit:  audit,
		logger: logger,
	}
}

func (m *AuditMiddleware) Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		arw := &auditResponseWriter{ResponseWriter: w}
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			m.logger.Error("Failed to read request body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body.Close()
		// steal the request body
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		next.ServeHTTP(arw, r)
		if arw.StatusCode == 200 {
			var metrics interface{}
			err := json.Unmarshal(requestBody, &metrics)
			if err != nil {
				m.logger.Error("Failed to unmarshal audit metrics", zap.Error(err))
				return
			}
			// notify dispatcher about new event
			m.audit.Notify(metrics, r.RemoteAddr)
		}
	})
}
