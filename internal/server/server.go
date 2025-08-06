package server

import (
	"log"
	"net/http"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/handler"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/middleware"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/repository"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
)

type Server struct {
	server *http.Server
}

func (s *Server) Run() error {
	log.Printf("Starting server on %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

func NewServer() *Server {
	var apiPrefix = "/update/"
	// repositories
	metricRepo := repository.NewMetricRepository()

	// services
	metricService := service.NewMetricService(metricRepo)

	// handlers
	metricHandler := handler.NewMetricHandler(metricService)

	// middleware
	contentTypeMiddleware := middleware.NewContentTypeMiddleware()

	// HTTP server setup
	mux := http.NewServeMux()
	mux.Handle(
		apiPrefix,
		contentTypeMiddleware.CheckContentType(
			http.HandlerFunc(metricHandler.HandleMetric),
		),
	)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return &Server{
		server: server,
	}
}
