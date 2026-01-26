package server

import (
	"context"
	"net/http"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/handler"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/middleware"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type httpServer struct {
	server *http.Server
	logger *zap.Logger
}

func NewHTTPServer(opts baseServerOpts) *httpServer {
	// routing
	r := chi.NewRouter()
	r.Use(middleware.HTTPLogMiddleware(opts.logger))
	if *opts.vars.TrustedSubnet != "" {
		r.Use(middleware.CheckCIDR(*opts.vars.TrustedSubnet))
	}
	r.Mount("/debug/pprof/", http.DefaultServeMux)

	// handlers
	metricHandler := handler.NewMetricHandler(opts.services.metricsService)

	// register metrics entries
	metricHandler.Register(r)

	httpSrv := &http.Server{
		Addr:    *opts.vars.Endpoint,
		Handler: r,
	}

	return &httpServer{
		server: httpSrv,
		logger: opts.logger,
	}
}

func (h *httpServer) Run() error {
	h.logger.Info("Starting HTTP server", zap.String("addr", h.server.Addr))

	return h.server.ListenAndServe()
}

func (h *httpServer) Shutdown() error {
	return h.server.Shutdown(context.TODO())
}
