package handler

import (
	"log"
	"net/http"
	"strings"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/go-chi/chi"
)

// SetMetric handles HTTP requests for updating a single metric via
// URL parameters. It extracts the metric name, value, and type from
// the request path, delegates the update to the service layer, and
// returns an appropriate HTTP status based on the outcome.
//
// @Summary      Update metric
// @Description  Updates a single metric identified by type and name using a value from the URL
// @Tags         metrics
// @Produce      plain
// @Param        type   path      string  true  "Metric type (counter or gauge)"
// @Param        name   path      string  true  "Metric name"
// @Param        value  path      string  true  "Metric value"
// @Success      200    {string}  string  "Metric updated"
// @Failure      400    {string}  string  "Invalid metric parameters or value"
// @Failure      500    {string}  string  "Internal Server Error"
// @Router       /update/{type}/{name}/{value} [post]
func (h *metricHandler) SetMetric(w http.ResponseWriter, r *http.Request) {
	metricName := strings.TrimSpace(chi.URLParam(r, "name"))
	metricValue := strings.TrimSpace(chi.URLParam(r, "value"))
	metricType := strings.TrimSpace(chi.URLParam(r, "type"))
	w.Header().Set("Content-Type", "text/plain")
	var err error
	switch metricType {
	case models.Gauge:
		err = h.service.SetGauge(metricName, metricValue)
	case models.Counter:
		err = h.service.SetCounter(metricName, metricValue)
	default:
		log.Printf("Unknown metric type: %s\n", metricType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("Error setting metric %s with value = %v, err: %v\n", metricName, metricValue, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Successfully set metric %s to %s\n", metricName, metricValue)
	w.WriteHeader(http.StatusOK)
}
