package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
)

func (h *metricHandler) SetMetricByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var metricErr *service.InvalidMetricError
	m, err := h.service.SetMetricByModel(body)
	if errors.As(err, &metricErr) {
		w.WriteHeader(metricErr.StatusCode)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
