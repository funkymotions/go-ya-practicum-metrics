package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
)

func (h *metricHandler) SetMetricBulk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var metricErr *service.InvalidMetricError
	hash := r.Header.Get("hashsha256")
	err = h.service.SetMetricBulk(body, []byte(hash))
	if errors.As(err, &metricErr) {
		w.WriteHeader(metricErr.StatusCode)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
	w.WriteHeader(http.StatusOK)
}
