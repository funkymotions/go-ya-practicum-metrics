package handler

import "net/http"

// GetAllMetrics handles HTTP requests for listing all metrics.
// It renders the metrics as HTML and writes them to the response.
//
// @Summary      Get all metrics
// @Description  Returns all metrics rendered as an HTML page
// @Tags         metrics
// @Produce      html
// @Success      200  {string}  string  "HTML page with all metrics"
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       / [get]
func (h *metricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.service.GetAllMetricsForHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(metrics))
	w.WriteHeader(http.StatusOK)
}
