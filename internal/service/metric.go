package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/repository"
)

type metricRepoInterface interface {
	SetGauge(name string, parameter float64)
	SetCounter(name string, parameter int64)
	SetGaugeIntrospect(name string, parameter float64) error
	SetCounterIntrospect(name string, parameter int64) error
	GetMetric(name string, metricType string) (*models.Metrics, bool)
	GetAllMetrics() map[string]models.Metrics
	SetMetricBulk(m *[]models.Metrics) error
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
	m, res := s.repo.GetMetric(name, metricType)
	return m, res
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
		// err := s.repo.SetGaugeIntrospect(metric.ID, *metric.Value)
		err := withRetry(func() error {
			return s.repo.SetGaugeIntrospect(metric.ID, *metric.Value)
		}, 0, 3)
		if err != nil {
			var nonRetriable *repository.NonRetriablePgError
			if errors.As(err, &nonRetriable) {
				return nil
			}
		}
		return err
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("invalid counter metric: %s", metric.ID)
		}
		err := withRetry(func() error {
			return s.repo.SetCounterIntrospect(metric.ID, *metric.Delta)
		}, 0, 3)
		if err != nil {
			var nonRetriable *repository.NonRetriablePgError
			if errors.As(err, &nonRetriable) {
				return nil
			}
		}
		return err
	default:
		return fmt.Errorf("invalid metric type: %s", metric.MType)
	}
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

func (s *metricService) SetMetricBulk(m *[]models.Metrics) error {
	return s.repo.SetMetricBulk(m)
}

func isMetricNameAlphanumeric(input string, r *regexp.Regexp) bool {
	return r.MatchString(input)
}

func withRetry(fn func() error, attempts int, maxAttempts int) error {
	if maxAttempts == 0 {
		return fn()
	}
	if attempts >= maxAttempts {
		return fmt.Errorf("max retry attempts reached")
	}
	secondsToSleep := time.Duration(2*(attempts)+1) * time.Second
	var retriableErr *repository.RetriablePgError
	if err := fn(); err != nil {
		if errors.As(err, &retriableErr) {
			time.Sleep(secondsToSleep)
			return withRetry(fn, attempts+1, maxAttempts)
		}
		return err
	}
	return nil
}
