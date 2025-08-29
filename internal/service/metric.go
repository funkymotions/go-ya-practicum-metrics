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
	GetMetric(name string, metricType string) (*models.Metrics, bool)
	GetAllMetrics() map[string]models.Metrics
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
	s.repo.SetCounter(name, value)
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
	s.repo.SetGauge(name, value)
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

func isMetricNameAlphanumeric(input string, r *regexp.Regexp) bool {
	return r.MatchString(input)
}
