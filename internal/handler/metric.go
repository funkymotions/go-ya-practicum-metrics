package handler

import (
	"net/http"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/middleware"
	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/go-chi/chi"
)

type metricService interface {
	SetCounter(name string, value string) error
	SetGauge(name string, value string) error
	SetMetricByModel(m *models.Metrics) error
	GetMetricByModel(m *models.Metrics) (*models.Metrics, error)
	GetMetric(metricType, name string) (*models.Metrics, bool)
	GetAllMetricsForHTML() string
	Ping() error
}

type metricHandler struct {
	service metricService
}

func NewMetricHandler(s metricService) *metricHandler {
	return &metricHandler{
		service: s,
	}
}

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
}
