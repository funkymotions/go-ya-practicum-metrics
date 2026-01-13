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
	SetMetricByModel([]byte) (*models.Metrics, error)
	GetMetricByModel(m *models.Metrics) (*models.Metrics, error)
	GetMetric(metricType, name string) (*models.Metrics, error)
	GetAllMetricsForHTML() string
	SetMetricBulk([]byte, []byte) error
	Ping() error
}

type auditMiddleware interface {
	Audit(next http.Handler) http.Handler
}

type metricHandler struct {
	service         metricService
	auditMiddleware auditMiddleware
}

func NewMetricHandler(s metricService, am auditMiddleware) *metricHandler {
	return &metricHandler{
		service:         s,
		auditMiddleware: am,
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
	engine.
		With(middleware.CompressHandler).
		With(h.auditMiddleware.Audit).
		Post("/updates/", http.HandlerFunc(h.SetMetricBulk))
}
