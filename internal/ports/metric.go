package ports

import (
	"github.com/funkymotions/go-ya-practicum-metrics/internal/dto"
	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
)

type MetricRepoReader interface {
	GetMetric(name string, metricType string) (*models.Metrics, bool)
	GetAllMetrics() map[string]models.Metrics
	Ping() error
}

type MetricRepoWriter interface {
	SetGauge(name string, parameter float64)
	SetCounter(name string, parameter int64)
	SetGaugeIntrospect(name string, parameter float64) error
	SetCounterIntrospect(name string, parameter int64) error
	SetMetricBulk(m *[]models.Metrics) error
}

type MetricServiceReader interface {
	GetMetricByModel(m *models.Metrics) (*models.Metrics, error)
	GetMetric(metricType, name string) (*models.Metrics, error)
	GetAllMetricsForHTML() string
	Ping() error
}

type MetricServiceWriter interface {
	SetCounter(name string, value string) error
	SetGauge(name string, value string) error
	SetMetricByModel([]byte) (*models.Metrics, error)
	SetMetricBulk([]models.Metrics, []byte, []byte, string) error
	SetEncryptedMetricBulk(dto.EncryptedMetrics, []byte, string) error
}
