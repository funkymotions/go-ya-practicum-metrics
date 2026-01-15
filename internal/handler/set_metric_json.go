package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
)

// SetMetricByJSON handles HTTP requests for updating a single metric using a
// JSON payload. It validates the Content-Type header, reads and forwards the
// body to the service layer, and writes the updated metric back as JSON or an
// appropriate error status.
//
// @Summary      Update metric by JSON
// @Description  Updates a single metric using a JSON payload
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Param        metric  body      models.Metrics  true  "Metric to update"
// @Success      200     {object}  models.Metrics  "Updated metric"
// @Failure      400     {string}  string          "Bad Request"
// @Failure      404     {string}  string          "Metric Not Found"
// @Failure      500     {string}  string          "Internal Server Error"
// @Router       /update/ [post]
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
