package agent

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"time"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"go.uber.org/zap"
)

const contentType = "application/json"

var names = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
	"NumGC",
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
			m.config.Logger.Info("Sending metrics to server...")
			m.mu.Lock()
			for name, metric := range m.metrics {
				b := prepareMetricBytes(&metric)
				m.config.Logger.Info("Sending metric", zap.String("name", name), zap.ByteString("body", b))
				r, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
				if err != nil {
					m.config.Logger.Error("Error creating request", zap.String("name", name), zap.Error(err))
					continue
				}
				r.Header.Set("Content-Type", contentType)
				r.Header.Set("Accept-Encoding", "gzip")
				resp, err := m.config.Client.Do(r)
				if err != nil {
					m.config.Logger.Error("Error sending metric", zap.String("name", name), zap.Error(err))
					continue
				}
				resp.Body.Close()
			}
			m.mu.Unlock()
		case <-stop:
			return
		}
	}
}

func (m *agent) collectMetrics(stop chan struct{}) {
	ticker := time.NewTicker(m.config.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.config.Logger.Info("Collecting metrics...")
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			m.mu.Lock()
			for _, name := range names {
				// using reflection to gather memStats metrics
				m.metrics[name] = getGaugeMetric(memStats, name)
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

func getGaugeMetric(stats runtime.MemStats, name string) models.Metrics {
	reflectValue := reflect.ValueOf(stats)
	field := reflectValue.FieldByName(name)
	fieldValue := field.Interface()
	var floatVal float64
	switch v := fieldValue.(type) {
	case uint64:
		floatVal = float64(v)
	case uint32:
		floatVal = float64(v)
	case float64:
		floatVal = v
	}
	return models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &floatVal,
	}
}

func prepareMetricBytes(m *models.Metrics) []byte {
	jsonData, _ := json.Marshal(m)
	return jsonData
}
