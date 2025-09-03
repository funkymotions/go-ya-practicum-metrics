package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type customResponseWriter struct {
	http.ResponseWriter
	StatusCode    uint
	Size          int
	isHeadersSent bool
}

func (w *customResponseWriter) WriteHeader(statusCode int) {
	if w.isHeadersSent {
		return
	}
	w.StatusCode = uint(statusCode)
	w.ResponseWriter.WriteHeader(statusCode)
	w.isHeadersSent = true
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	if !w.isHeadersSent {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.Size += n
	return n, err
}

func HTTPLogMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			crw := &customResponseWriter{
				ResponseWriter: w,
			}
			uri := r.RequestURI
			method := r.Method
			start := time.Now()
			logger.Info("New HTTP request",
				zap.String("method", method),
				zap.String("uri", uri),
			)
			next.ServeHTTP(crw, r)
			duration := time.Since(start)
			logger.Info("HTTP request finished",
				zap.String("method", method),
				zap.String("uri", uri),
				zap.Duration("elapsedTime", duration),
				zap.Uint("statusCode", crw.StatusCode),
				zap.Int("length", crw.Size),
			)
		})
	}
}
