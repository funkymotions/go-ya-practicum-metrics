package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/utils"
)

type InvalidMetricError struct {
	Message    string
	StatusCode int
}

func (e *InvalidMetricError) Error() string {
	return fmt.Sprintf("code %d: %s", e.StatusCode, e.Message)
}

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
	repo       metricRepoInterface
	re         *regexp.Regexp
	hashSecret []byte
}

func NewMetricService(repo metricRepoInterface, hashSecret []byte) *metricService {
	return &metricService{
		repo:       repo,
		re:         regexp.MustCompile(`^\w+$`),
		hashSecret: hashSecret,
	}
}

func (s *metricService) SetCounter(name string, rawValue string) error {
	if !isMetricNameAlphanumeric(name, s.re) {
		return &InvalidMetricError{
			Message:    fmt.Sprintf("invalid metric name: %s", name),
			StatusCode: http.StatusBadRequest,
		}
	}
	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return &InvalidMetricError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}
	s.repo.SetCounterIntrospect(name, value)
	return nil
}

func (s *metricService) SetGauge(name string, rawValue string) error {
	if !isMetricNameAlphanumeric(name, s.re) {
		return &InvalidMetricError{
			Message:    fmt.Sprintf("invalid metric name: %s", name),
			StatusCode: http.StatusBadRequest,
		}
	}
	value, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return &InvalidMetricError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}
	s.repo.SetGaugeIntrospect(name, value)
	return nil
}

func (s *metricService) GetMetric(name string, metricType string) (*models.Metrics, error) {
	if !isMetricNameAlphanumeric(name, s.re) {
		return nil, &InvalidMetricError{
			Message:    fmt.Sprintf("invalid metric name: %s", name),
			StatusCode: http.StatusBadRequest,
		}
	}
	m, res := s.repo.GetMetric(name, metricType)
	if !res {
		return nil, &InvalidMetricError{
			Message:    fmt.Sprintf("metric not found: %s", name),
			StatusCode: http.StatusNotFound,
		}
	}
	return m, nil
}

func (s *metricService) GetAllMetricsForHTML() string {
	metrics := s.repo.GetAllMetrics()
	var result string
	for _, m := range metrics {
		result += fmt.Sprintf("%s\n", m.String())
	}
	return result
}

func (s *metricService) SetMetricByModel(input []byte) (*models.Metrics, error) {
	var metric models.Metrics
	if err := json.NewDecoder(bytes.NewReader(input)).Decode(&metric); err != nil {
		return nil, &InvalidMetricError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}
	if !isMetricNameAlphanumeric(metric.ID, s.re) {
		return nil, &InvalidMetricError{
			Message:    fmt.Sprintf("invalid metric name: %s", metric.ID),
			StatusCode: http.StatusBadRequest,
		}
	}
	if !isMetricDataOK(&metric) {
		return nil, &InvalidMetricError{
			Message:    fmt.Sprintf("invalid metric data: %s", metric.ID),
			StatusCode: http.StatusBadRequest,
		}
	}
	var retriableFn func() error
	switch metric.MType {
	case models.Gauge:
		retriableFn = func() error {
			return s.repo.SetGaugeIntrospect(metric.ID, *metric.Value)
		}
	case models.Counter:
		retriableFn = func() error {
			return s.repo.SetCounterIntrospect(metric.ID, *metric.Delta)
		}
	default:
		return nil, &InvalidMetricError{
			Message:    fmt.Sprintf("invalid metric type: %s", metric.MType),
			StatusCode: http.StatusBadRequest,
		}
	}
	// TODO: make maxAttempts configurable
	if err := utils.WithRetry(retriableFn, 0, 3); err != nil {
		return nil, &InvalidMetricError{
			Message:    fmt.Sprintf("failed to set metric: %s", err.Error()),
			StatusCode: http.StatusInternalServerError,
		}
	}
	return &metric, nil
}

func (s *metricService) GetMetricByModel(metric *models.Metrics) (*models.Metrics, error) {
	if !isMetricNameAlphanumeric(metric.ID, s.re) {
		return nil, &InvalidMetricError{
			Message:    fmt.Sprintf("invalid metric name: %s", metric.ID),
			StatusCode: http.StatusBadRequest,
		}
	}
	m, found := s.repo.GetMetric(metric.ID, metric.MType)
	if !found {
		return nil, &InvalidMetricError{
			Message:    fmt.Sprintf("metric not found: %s", metric.ID),
			StatusCode: http.StatusNotFound,
		}
	}
	return m, nil
}

func (s *metricService) Ping() error {
	return s.repo.Ping()
}

func (s *metricService) SetMetricBulk(input []byte, signature []byte) error {
	if len(s.hashSecret) > 0 {
		if ok := isHashValid(signature, input, s.hashSecret); !ok {
			return &InvalidMetricError{
				Message:    "invalid hash",
				StatusCode: http.StatusBadRequest,
			}
		}
	}
	var metrics []models.Metrics
	if err := json.NewDecoder(bytes.NewReader(input)).Decode(&metrics); err != nil {
		return &InvalidMetricError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}
	return s.repo.SetMetricBulk(&metrics)
}

func isMetricNameAlphanumeric(input string, r *regexp.Regexp) bool {
	return r.MatchString(input)
}

func isMetricDataOK(m *models.Metrics) bool {
	switch m.MType {
	case models.Gauge:
		return m.Value != nil
	case models.Counter:
		return m.Delta != nil
	default:
		return false
	}
}

func isHashValid(signature, payload, secret []byte) bool {
	if len(signature) == 0 || len(payload) == 0 || len(secret) == 0 {
		return false
	}
	decodedSignature := make([]byte, sha256.Size)
	_, err := hex.Decode(decodedSignature, signature)
	if err != nil {
		return false
	}
	h := hmac.New(sha256.New, secret)
	h.Write(payload)
	hash := h.Sum(nil)
	return hmac.Equal(decodedSignature, hash)
}
