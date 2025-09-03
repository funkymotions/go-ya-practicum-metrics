package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/go-chi/chi"
)

type metricService interface {
	SetCounter(name string, value string) error
	SetGauge(name string, value string) error
	SetMetricByModel(m *models.Metrics) error
	GetMetricByModel(m *models.Metrics) (*models.Metrics, error)
	GetMetric(metricType, name string) (*models.Metrics, bool)
	GetAllMetricsForHTML() string
}

type metricHandler struct {
	service metricService
}

func NewMetricHandler(s metricService) *metricHandler {
	return &metricHandler{
		service: s,
	}
}

func (h *metricHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metricName := strings.TrimSpace(chi.URLParam(r, "name"))
	metricType := strings.TrimSpace(chi.URLParam(r, "type"))
	metric, found := h.service.GetMetric(metricName, metricType)
	w.Header().Set("Content-Type", "text/plain")
	if !found {
		log.Printf("Metric %s not found\n", metricName)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write([]byte(metric.String()))
}

func (h *metricHandler) SetMetric(w http.ResponseWriter, r *http.Request) {
	metricName := strings.TrimSpace(chi.URLParam(r, "name"))
	metricValue := strings.TrimSpace(chi.URLParam(r, "value"))
	metricType := strings.TrimSpace(chi.URLParam(r, "type"))
	w.Header().Set("Content-Type", "text/plain")
	var err error
	switch metricType {
	case models.Gauge:
		err = h.service.SetGauge(metricName, metricValue)
	case models.Counter:
		err = h.service.SetCounter(metricName, metricValue)
	default:
		log.Printf("Unknown metric type: %s\n", metricType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("Error setting metric %s with value = %v, err: %v\n", metricName, metricValue, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Successfully set metric %s to %s\n", metricName, metricValue)
	w.WriteHeader(http.StatusOK)
}

func (h *metricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.service.GetAllMetricsForHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(metrics))
	w.WriteHeader(http.StatusOK)
}

func (h *metricHandler) SetMetricByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.service.SetMetricByModel(&metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *metricHandler) GetMetricByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m, err := h.service.GetMetricByModel(&metric)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
