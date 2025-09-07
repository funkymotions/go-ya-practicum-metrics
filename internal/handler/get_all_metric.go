package handler

import "net/http"

func (h *metricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.service.GetAllMetricsForHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(metrics))
	w.WriteHeader(http.StatusOK)
}
