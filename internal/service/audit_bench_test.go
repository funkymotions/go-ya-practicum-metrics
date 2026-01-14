package service

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
)

func BenchmarkExtractMetricNamesSingle(b *testing.B) {
	m := models.Metrics{ID: "A", MType: models.Gauge}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractMetricNames(m)
	}
}

func BenchmarkExtractMetricNamesSlice(b *testing.B) {
	ms := make([]models.Metrics, 1000)
	for i := range ms {
		ms[i] = models.Metrics{ID: "M", MType: models.Gauge}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractMetricNames(ms)
	}
}

func BenchmarkAuditWriteEvent(b *testing.B) {
	f, err := os.CreateTemp("", "audit-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	s := &auditService{filePath: f.Name()}
	msg := AuditMessage{Timestamp: time.Now().Unix(), Metrics: []string{"id"}, IPAddress: "127.0.0.1"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.writeEvent(msg)
	}
}

func BenchmarkAuditSendEvent(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	s := &auditService{remoteURL: srv.URL}
	msg := AuditMessage{Timestamp: time.Now().Unix(), Metrics: []string{"id"}, IPAddress: "127.0.0.1"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.sendEvent(msg)
	}
}
