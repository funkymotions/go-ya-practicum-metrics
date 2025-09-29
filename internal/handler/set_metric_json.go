package handler

import (
	"encoding/json"
	"net/http"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
)

func (h *metricHandler) SetMetricByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	contentType := r.Header.Get("Content-Type")
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
