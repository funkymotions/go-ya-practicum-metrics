package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

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
