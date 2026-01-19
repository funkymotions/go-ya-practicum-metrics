package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
)

// GetMetricByJSON handles HTTP requests for retrieving a metric using a JSON payload.
// It validates the request content type, decodes the metric from the body,
// fetches the corresponding metric from the service layer, and writes it as JSON.
//
// @Summary      Get metric by JSON
// @Description  Returns a single metric identified in the JSON payload
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Param        metric  body      models.Metrics  true  "Metric descriptor"
// @Success      200     {object}  models.Metrics  "Metric found"
// @Failure      400     {string}  string          "Bad Request"
// @Failure      404     {string}  string          "Metric Not Found"
// @Failure      500     {string}  string          "Internal Server Error"
// @Router       /value/ [post]
func (h *metricHandler) GetMetricByJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO: move deserialization to service layer
	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m, err := h.service.GetMetricByModel(&metric)
	var metricErr *service.InvalidMetricError
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
}
