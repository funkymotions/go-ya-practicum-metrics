package ports

import models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"

type MetricRepoInterface interface {
	SetGauge(name string, parameter float64)
	SetCounter(name string, parameter int64)
	SetGaugeIntrospect(name string, parameter float64) error
	SetCounterIntrospect(name string, parameter int64) error
	GetMetric(name string, metricType string) (*models.Metrics, bool)
	GetAllMetrics() map[string]models.Metrics
	SetMetricBulk(m *[]models.Metrics) error
	Ping() error
}

type MetricService interface {
	SetCounter(name string, value string) error
	SetGauge(name string, value string) error
	SetMetricByModel([]byte) (*models.Metrics, error)
	GetMetricByModel(m *models.Metrics) (*models.Metrics, error)
	GetMetric(metricType, name string) (*models.Metrics, error)
	GetAllMetricsForHTML() string
	SetMetricBulk([]byte, []byte, string) error
	SetEncryptedMetricBulk(input []byte, signature []byte, remoteIP string) error
	Ping() error
}
