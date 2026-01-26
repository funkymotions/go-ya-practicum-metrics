package handler

import (
	"net/http"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/middleware"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/ports"
	"github.com/go-chi/chi"
)

type metricHandler struct {
	service ports.MetricService
}

func NewMetricHandler(s ports.MetricService) *metricHandler {
	return &metricHandler{
		service: s,
	}
}

// Register configures the HTTP routes for working with metrics on the
// provided chi.Mux router, including health checks, HTML rendering,
// plain-text endpoints, and JSON-based update/value APIs.
func (h *metricHandler) Register(engine *chi.Mux) {
	engine.Get("/ping", h.Ping)
	engine.
		With(middleware.CompressHandler).
		Get("/", http.HandlerFunc(h.GetAllMetrics))
	engine.Get("/value/{type}/{name}", http.HandlerFunc(h.GetMetric))
	engine.Post("/update/{type}/{name}/{value}", http.HandlerFunc(h.SetMetric))
	engine.
		With(middleware.CompressHandler).
		Post("/update/", http.HandlerFunc(h.SetMetricByJSON))
	engine.
		With(middleware.CompressHandler).
		Post("/value/", http.HandlerFunc(h.GetMetricByJSON))
	engine.
		With(middleware.CompressHandler).
		Post("/updates/", http.HandlerFunc(h.SetMetricBulk))
}
