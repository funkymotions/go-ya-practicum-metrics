package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi"
)

// GetMetric handles HTTP requests for retrieving a single metric by type and name.
// It extracts parameters from the URL, delegates lookup to the service layer,
// and writes the metric value as plain text or an appropriate error status.
//
// @Summary      Get metric
// @Description  Returns the current value of a single metric by type and name
// @Tags         metrics
// @Produce      plain
// @Param        type  path      string  true  "Metric type (counter or gauge)"
// @Param        name  path      string  true  "Metric name"
// @Success      200   {string}  string  "Metric value"
// @Failure      400   {string}  string  "Invalid metric parameters"
// @Failure      404   {string}  string  "Metric not found"
// @Failure      500   {string}  string  "Internal Server Error"
// @Router       /value/{type}/{name} [get]
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
