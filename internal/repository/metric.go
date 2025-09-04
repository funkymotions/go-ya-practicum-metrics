package repository

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
)

type metricRepository struct {
	memStorage    map[string]models.Metrics
	mu            sync.RWMutex
	writeInterval time.Duration
	filePath      string
}

func NewMetricRepository(filePath string, isRestoreNeeded bool, writeInterval time.Duration) *metricRepository {
	r := &metricRepository{
		memStorage:    make(map[string]models.Metrics),
		mu:            sync.RWMutex{},
		writeInterval: writeInterval,
		filePath:      filePath,
	}
	if isRestoreNeeded {
		r.readMetricsFromFile()
	}
	if writeInterval > 0 && filePath != "" {
		ticker := time.NewTicker(writeInterval)
		go func() {
			for range ticker.C {
				r.writeMetricsToFile()
			}
		}()
	}
	return r
}

func (r *metricRepository) SetGauge(name string, value float64) {
	r.mu.Lock()
	key := models.Gauge + ":" + name
	m, exists := r.memStorage[key]
	if !exists {
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
			Hash:  "",
		}
		r.memStorage[key] = metric
	} else {
		m.Value = &value
		r.memStorage[key] = m
	}
	r.mu.Unlock()
	// write metrics to disk in same request goroutine
	if r.writeInterval == 0 {
		r.writeMetricsToFile()
	}
}

func (r *metricRepository) SetCounter(name string, delta int64) {
	r.mu.Lock()
	key := models.Counter + ":" + name
	m, exists := r.memStorage[key]
	if !exists {
		metric := models.Metrics{
			ID:    name,
			MType: models.Counter,
			Value: nil,
			Delta: &delta,
			Hash:  "",
		}
		r.memStorage[key] = metric
	} else {
		*m.Delta += delta
		r.memStorage[key] = m
	}
	r.mu.Unlock()
	// write metrics to disk in same request goroutine
	if r.writeInterval == 0 {
		r.writeMetricsToFile()
	}
}

func (r *metricRepository) GetMetric(name string, metricType string) (*models.Metrics, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := metricType + ":" + name
	m, exists := r.memStorage[key]
	return &m, exists
}

func (r *metricRepository) GetAllMetrics() map[string]models.Metrics {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.memStorage
}

func (r *metricRepository) writeMetricsToFile() error {
	if len(r.memStorage) == 0 {
		return nil
	}
	f, err := os.OpenFile(r.filePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	defer buf.Flush()
	r.mu.RLock()
	defer r.mu.RUnlock()
	var metrics []models.Metrics
	for _, m := range r.memStorage {
		metrics = append(metrics, m)
	}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metrics); err != nil {
		return err
	}
	return nil
}

func (r *metricRepository) readMetricsFromFile() error {
	f, err := os.OpenFile(r.filePath, os.O_RDONLY, 0444)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	var metrics []models.Metrics
	if err := json.NewDecoder(buf).Decode(&metrics); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range metrics {
		key := m.MType + ":" + m.ID
		r.memStorage[key] = m
	}
	return nil
}
