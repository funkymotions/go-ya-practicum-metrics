package server

import (
	"log"
	"net/http"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/db"
	appenv "github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	sql "github.com/funkymotions/go-ya-practicum-metrics/internal/driver/db"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/handler"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/logger"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/middleware"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/repository"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type Server struct {
	server            *http.Server
	logger            *zap.Logger
	stopCh            chan struct{}
	doneCh            chan struct{}
	shouldWaitForDone bool
}

func (s *Server) Run() error {
	s.logger.Info("Starting server", zap.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown() {
	s.logger.Warn("Shutting down server", zap.String("addr", s.server.Addr))
	// notify all subscribed goroutines to exit
	close(s.stopCh)
	if s.shouldWaitForDone {
		<-s.doneCh
		s.logger.Info("All goroutines have exited")
	}
}

func NewServer(v *appenv.Variables) *Server {
	// db
	if v.DatabaseDSN == nil {
		log.Fatal("database dsn is not set")
	}
	dbConf := db.NewDbConfig(*v.DatabaseDSN)
	d, err := sql.NewSQLDriver(dbConf)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	// logger
	logger, err := logger.NewLogger(zap.NewAtomicLevelAt(zap.InfoLevel))
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	// channels
	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	// repositories
	metricRepo := repository.NewMetricRepository(
		*v.FileStoragePath,
		*v.Restore,
		time.Second*time.Duration(*v.StoreInterval),
		d,
		stopCh,
		doneCh,
	)
	// services
	metricService := service.NewMetricService(metricRepo)
	// handlers
	metricHandler := handler.NewMetricHandler(metricService)
	// routing
	r := chi.NewRouter()
	r.Use(middleware.HTTPLogMiddleware(logger))
	// register metrics entries
	metricHandler.Register(r)
	httpSrv := &http.Server{
		Addr:    *v.Endpoint,
		Handler: r,
	}
	return &Server{
		server:            httpSrv,
		logger:            logger,
		stopCh:            stopCh,
		doneCh:            doneCh,
		shouldWaitForDone: *v.StoreInterval != 0,
	}
}
