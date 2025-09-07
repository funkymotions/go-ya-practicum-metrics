package service

import (
	"fmt"
	"regexp"
	"strconv"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
)

type metricRepoInterface interface {
	SetGauge(name string, parameter float64)
	SetCounter(name string, parameter int64)
	SetGaugeIntrospect(name string, parameter float64)
	SetCounterIntrospect(name string, parameter int64)
	GetMetric(name string, metricType string) (*models.Metrics, bool)
	GetAllMetrics() map[string]models.Metrics
	Ping() error
}

type metricService struct {
	repo metricRepoInterface
	re   *regexp.Regexp
}

func NewMetricService(repo metricRepoInterface) *metricService {
	return &metricService{
		repo: repo,
		re:   regexp.MustCompile(`^\w+$`),
	}
}

func (s *metricService) SetCounter(name string, rawValue string) error {
	if !isMetricNameAlphanumeric(name, s.re) {
		return fmt.Errorf("invalid metric name: %s", name)
	}
	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return err
	}
	s.repo.SetCounterIntrospect(name, value)
	return nil
}

func (s *metricService) SetGauge(name string, rawValue string) error {
	if !isMetricNameAlphanumeric(name, s.re) {
		return fmt.Errorf("invalid metric name: %s", name)
	}
	value, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return err
	}
	s.repo.SetGaugeIntrospect(name, value)
	return nil
}

func (s *metricService) GetMetric(name string, metricType string) (*models.Metrics, bool) {
	if !isMetricNameAlphanumeric(name, s.re) {
		return nil, false
	}
	return s.repo.GetMetric(name, metricType)
}

func (s *metricService) GetAllMetricsForHTML() string {
	metrics := s.repo.GetAllMetrics()
	var result string
	for _, m := range metrics {
		result += fmt.Sprintf("%s\n", m.String())
	}
	return result
}

func (s *metricService) SetMetricByModel(metric *models.Metrics) error {
	if !isMetricNameAlphanumeric(metric.ID, s.re) {
		return fmt.Errorf("invalid metric name: %s", metric.ID)
	}
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("invalid gauge metric: %s", metric.ID)
		}
		s.repo.SetGaugeIntrospect(metric.ID, *metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("invalid counter metric: %s", metric.ID)
		}
		s.repo.SetCounterIntrospect(metric.ID, *metric.Delta)
	default:
		return fmt.Errorf("invalid metric type: %s", metric.MType)
	}
	return nil
}

func (s *metricService) GetMetricByModel(metric *models.Metrics) (*models.Metrics, error) {
	if !isMetricNameAlphanumeric(metric.ID, s.re) {
		return nil, fmt.Errorf("invalid metric name: %s", metric.ID)
	}
	m, found := s.repo.GetMetric(metric.ID, metric.MType)
	if !found {
		return nil, fmt.Errorf("metric not found: %s", metric.ID)
	}
	return m, nil
}

func (s *metricService) Ping() error {
	return s.repo.Ping()
}

func isMetricNameAlphanumeric(input string, r *regexp.Regexp) bool {
	return r.MatchString(input)
}
