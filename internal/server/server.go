package server

import (
	"log"
	"net/http"

	appenv "github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/handler"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/logger"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/middleware"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/repository"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type Server struct {
	server *http.Server
	logger *zap.Logger
}

func (s *Server) Run() error {
	s.logger.Info("Starting server", zap.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

func NewServer(v *appenv.Variables) *Server {
	var apiPrefix = "/update"
	// logger
	logger, err := logger.NewLogger(zap.NewAtomicLevelAt(zap.InfoLevel))
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	// repositories
	metricRepo := repository.NewMetricRepository()
	// services
	metricService := service.NewMetricService(metricRepo)
	// handlers
	metricHandler := handler.NewMetricHandler(metricService)
	// routing
	r := chi.NewRouter()
	r.Use(middleware.HTTPLogMiddleware(logger))
	r.Get("/", http.HandlerFunc(metricHandler.GetAllMetrics))
	r.Get("/value/{type}/{name}", http.HandlerFunc(metricHandler.GetMetric))
	r.Post(apiPrefix+"/{type}/{name}/{value}", http.HandlerFunc(metricHandler.SetMetric))
	server := &http.Server{
		Addr:    v.Endpoint,
		Handler: r,
	}
	return &Server{
		server: server,
		logger: logger,
	}
}
