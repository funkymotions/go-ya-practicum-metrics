package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi"
)

func (h *metricHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metricName := strings.TrimSpace(chi.URLParam(r, "name"))
	metricType := strings.TrimSpace(chi.URLParam(r, "type"))
	metric, err := h.service.GetMetric(metricName, metricType)
	w.Header().Set("Content-Type", "text/plain")
	var metricErr *service.InvalidMetricError
	if errors.As(err, &metricErr) {
		log.Printf("error while searching metric: %s, %s\n", metricName, metricErr)
		w.WriteHeader(metricErr.StatusCode)
		return
	}
	if err != nil {
		log.Printf("error while searching metric: %s, %v\n", metricName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte(metric.String()))
}
