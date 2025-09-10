package handler

import (
	"encoding/json"
	"net/http"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
)

func (h *metricHandler) SetMetricBulk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var metrics []models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.service.SetMetricBulk(&metrics); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
	w.WriteHeader(http.StatusOK)
}
