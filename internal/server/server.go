package server

import (
	"log"
	"net/http"

	appenv "github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/handler"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/repository"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi"
)

type Server struct {
	server *http.Server
}

func (s *Server) Run() error {
	log.Printf("Starting server on %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

func NewServer(v *appenv.Variables) *Server {
	var apiPrefix = "/update"
	// repositories
	metricRepo := repository.NewMetricRepository()
	// services
	metricService := service.NewMetricService(metricRepo)
	// handlers
	metricHandler := handler.NewMetricHandler(metricService)
	// routing
	r := chi.NewRouter()
	r.Get("/", http.HandlerFunc(metricHandler.GetAllMetrics))
	r.Get("/value/{type}/{name}", http.HandlerFunc(metricHandler.GetMetric))
	r.Post(apiPrefix+"/{type}/{name}/{value}", http.HandlerFunc(metricHandler.SetMetric))
	server := &http.Server{
		Addr:    v.Endpoint,
		Handler: r,
	}
	return &Server{
		server: server,
	}
}
