package repository

import (
	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
)

type metricRepository struct {
	memStorage map[string]models.Metrics
}

func NewMetricRepository() *metricRepository {
	return &metricRepository{
		memStorage: make(map[string]models.Metrics),
	}
}

func (r *metricRepository) SetGauge(name string, value float64) {
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
}

func (r *metricRepository) SetCounter(name string, delta int64) {
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
}

func (r *metricRepository) GetMetric(name string, metricType string) (*models.Metrics, bool) {
	key := metricType + ":" + name
	m, exists := r.memStorage[key]
	return &m, exists
}

func (r *metricRepository) GetAllMetrics() map[string]models.Metrics {
	return r.memStorage
}
