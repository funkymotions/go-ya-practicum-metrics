package handler

import (
	"log"
	"net/http"
	"strings"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
)

type metricService interface {
	SetCounter(name string, value string) error
	SetGauge(name string, value string) error
}

type metricHandler struct {
	service metricService
}

func NewMetricHandler(s metricService) *metricHandler {
	return &metricHandler{
		service: s,
	}
}

func (h *metricHandler) HandleMetric(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		log.Printf("Invalid path: %s\n", path)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType, metricName, metricValue := parts[1], parts[2], parts[3]
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Invalid method %s for path: %s\n", r.Method, path)
		return
	}
	var err error
	switch metricType {
	case models.Counter:
		err = h.service.SetCounter(metricName, metricValue)
	case models.Gauge:
		err = h.service.SetGauge(metricName, metricValue)
	default:
		log.Printf("Unknown metric type: %s\n", metricType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error setting %s metric: %v\n", metricType, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	log.Printf("%s metric %s successfully updated with %s\n", metricType, metricName, metricValue)
	w.WriteHeader(http.StatusOK)
}
