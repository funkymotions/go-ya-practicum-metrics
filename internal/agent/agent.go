package agent

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/utils"
	"go.uber.org/zap"
)

type getter func(runtime.MemStats) float64

const contentType = "application/json"

var getters = map[string]getter{
	"Alloc":         func(stats runtime.MemStats) float64 { return float64(stats.Alloc) },
	"BuckHashSys":   func(stats runtime.MemStats) float64 { return float64(stats.BuckHashSys) },
	"Frees":         func(stats runtime.MemStats) float64 { return float64(stats.Frees) },
	"GCCPUFraction": func(stats runtime.MemStats) float64 { return float64(stats.GCCPUFraction) },
	"GCSys":         func(stats runtime.MemStats) float64 { return float64(stats.GCSys) },
	"HeapAlloc":     func(stats runtime.MemStats) float64 { return float64(stats.HeapAlloc) },
	"HeapIdle":      func(stats runtime.MemStats) float64 { return float64(stats.HeapIdle) },
	"HeapInuse":     func(stats runtime.MemStats) float64 { return float64(stats.HeapInuse) },
	"HeapObjects":   func(stats runtime.MemStats) float64 { return float64(stats.HeapObjects) },
	"HeapReleased":  func(stats runtime.MemStats) float64 { return float64(stats.HeapReleased) },
	"HeapSys":       func(stats runtime.MemStats) float64 { return float64(stats.HeapSys) },
	"LastGC":        func(stats runtime.MemStats) float64 { return float64(stats.LastGC) },
	"Lookups":       func(stats runtime.MemStats) float64 { return float64(stats.Lookups) },
	"MCacheInuse":   func(stats runtime.MemStats) float64 { return float64(stats.MCacheInuse) },
	"MCacheSys":     func(stats runtime.MemStats) float64 { return float64(stats.MCacheSys) },
	"Mallocs":       func(stats runtime.MemStats) float64 { return float64(stats.Mallocs) },
	"NextGC":        func(stats runtime.MemStats) float64 { return float64(stats.NextGC) },
	"NumForcedGC":   func(stats runtime.MemStats) float64 { return float64(stats.NumForcedGC) },
	"OtherSys":      func(stats runtime.MemStats) float64 { return float64(stats.OtherSys) },
	"PauseTotalNs":  func(stats runtime.MemStats) float64 { return float64(stats.PauseTotalNs) },
	"MSpanInuse":    func(stats runtime.MemStats) float64 { return float64(stats.MSpanInuse) },
	"MSpanSys":      func(stats runtime.MemStats) float64 { return float64(stats.MSpanSys) },
	"StackInuse":    func(stats runtime.MemStats) float64 { return float64(stats.StackInuse) },
	"StackSys":      func(stats runtime.MemStats) float64 { return float64(stats.StackSys) },
	"Sys":           func(stats runtime.MemStats) float64 { return float64(stats.Sys) },
	"TotalAlloc":    func(stats runtime.MemStats) float64 { return float64(stats.TotalAlloc) },
	"NumGC":         func(stats runtime.MemStats) float64 { return float64(stats.NumGC) },
}

type agent struct {
	config  *Config
	metrics map[string]models.Metrics
	mu      sync.Mutex
}

type Config struct {
	Client         *http.Client
	PollInterval   time.Duration
	ReportInterval time.Duration
	MetricURL      url.URL
	Logger         *zap.Logger
	MaxRetries     int
	Hashing        struct {
		Key        *string
		HeaderName string
	}
}

type retriableError struct {
	err error
}

func newRetriableError(err error) *retriableError {
	return &retriableError{err: err}
}

func (r *retriableError) Error() string {
	return r.err.Error()
}

func (r *retriableError) Unwrap() error {
	return r.err
}

func NewAgent(cfg *Config) *agent {
	return &agent{
		config:  cfg,
		metrics: make(map[string]models.Metrics),
	}
}

func (m *agent) Launch() {
	m.config.Logger.Info(
		"Starting agent:",
		zap.String("metricURL", m.config.MetricURL.String()),
		zap.Duration("reportInterval", m.config.ReportInterval),
		zap.Duration("pollInterval", m.config.PollInterval),
	)
	stop := make(chan struct{})
	go m.collectMetrics(stop)
	go m.sendMetrics(stop)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	close(stop)
}

func (m *agent) sendMetrics(stop chan struct{}) {
	url := m.config.MetricURL.String()
	ticker := time.NewTicker(m.config.ReportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			utils.WithRetry(func() error {
				return m.performRequest(url)
			}, 0, m.config.MaxRetries)
		case <-stop:
			return
		}
	}
}

func hashBodyByKey(key *string, body []byte) string {
	hmac := hmac.New(sha256.New, []byte(*key))
	hmac.Write(body)
	return hex.EncodeToString(hmac.Sum(nil))
}

func (m *agent) performRequest(url string) (err error) {
	defer m.mu.Unlock()
	m.config.Logger.Info("Sending metrics to server...")
	m.mu.Lock()
	body := prepareRequestBody(m.metrics)
	m.config.Logger.Info("Sending metrics", zap.ByteString("body", body))
	r, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		m.config.Logger.Error("Error creating request", zap.Error(err))
		return newRetriableError(err)
	}
	r.Header.Set("Content-Type", contentType)
	r.Header.Set("Accept-Encoding", "gzip")
	if m.config.Hashing.Key != nil && *m.config.Hashing.Key != "" {
		hValue := hashBodyByKey(m.config.Hashing.Key, body)
		r.Header.Set(m.config.Hashing.HeaderName, hValue)
	}
	resp, err := m.config.Client.Do(r)
	if err != nil {
		m.config.Logger.Error("Error sending metrics", zap.Error(err))
		return newRetriableError(err)
	}
	if resp.StatusCode != http.StatusOK {
		m.config.Logger.Error("Non-OK HTTP status", zap.Int("status", resp.StatusCode))
		return newRetriableError(fmt.Errorf("non-OK HTTP status: %s", resp.Status))
	}
	resp.Body.Close()
	return nil
}

func (m *agent) collectMetrics(stop chan struct{}) {
	ticker := time.NewTicker(m.config.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			m.mu.Lock()
			m.config.Logger.Info("Collecting metrics...")
			for key, getter := range getters {
				m.metrics[key] = getGaugeMetricModel(key, memStats, getter)
			}
			randVal := float64(rand.Intn(1000))
			m.metrics["RandomValue"] = models.Metrics{
				ID:    "RandomValue",
				MType: models.Gauge,
				Value: &randVal,
			}
			pCount, ok := m.metrics["PollCount"]
			if ok {
				*pCount.Delta += 1
				m.metrics["PollCount"] = pCount
			} else {
				var initVal int64 = 1
				m.metrics["PollCount"] = models.Metrics{
					ID:    "PollCount",
					MType: models.Counter,
					Delta: &initVal,
				}
			}
			m.mu.Unlock()
		case <-stop:
			return
		}
	}
}

func getGaugeMetricModel(name string, stats runtime.MemStats, g getter) models.Metrics {
	value := g(stats)
	return models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	}
}

func prepareRequestBody(m map[string]models.Metrics) []byte {
	var metrics []models.Metrics
	for _, metric := range m {
		metrics = append(metrics, metric)
	}
	jsonData, _ := json.Marshal(metrics)
	return jsonData
}
