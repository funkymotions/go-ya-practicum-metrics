package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
)

// SetMetricBulk handles HTTP requests for updating multiple metrics in a single
// JSON payload. It validates the Content-Type header, reads the request body,
// delegates bulk update logic to the service layer, and writes an empty JSON
// object on success or an appropriate error status code.
//
// @Summary      Bulk update metrics
// @Description  Updates multiple metrics in a single JSON request
// @Tags         metrics
// @Accept       json
// @Produce      json
// @Param        hashsha256  header    string            false  "HMAC-SHA256 hash of payload"
// @Param        metrics     body      []models.Metrics  true   "List of metrics to update"
// @Success      200         {object}  map[string]any    "Empty JSON object on success"
// @Failure      400         {string}  string            "Bad Request"
// @Failure      404         {string}  string            "Metric Not Found"
// @Failure      500         {string}  string            "Internal Server Error"
// @Router       /updates/ [post]
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
