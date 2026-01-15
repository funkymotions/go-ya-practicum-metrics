package handler

import "net/http"

// Ping handles health-check requests for the metrics service.
// It delegates to the service Ping method and returns 200 OK on success
// or 500 Internal Server Error if the service is unavailable.
//
// @Summary      Health check
// @Description  Checks the availability of the metrics service
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "Service is healthy"
// @Failure      500  {string}  string  "Service is unavailable"
// @Router       /ping [get]
func (h metricHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
