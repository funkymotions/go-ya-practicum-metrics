package server

import (
	"crypto/rsa"
	"log"
	"time"

	_ "net/http/pprof"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/db"
	appenv "github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/driver"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/logger"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/ports"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/repository"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/service"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type services struct {
	metricsService interface {
		ports.MetricServiceWriter
		ports.MetricServiceReader
	}
}

type repositories struct {
	metricRepo interface {
		ports.MetricRepoWriter
		ports.MetricRepoReader
	}
}

type App struct {
	stopCh            chan struct{}
	doneCh            chan struct{}
	auditDoneCh       chan struct{}
	shouldWaitForDone bool
	opts              baseServerOpts
	*httpServer
	*grpcServer
}

type baseServerOpts struct {
	vars         *appenv.Variables
	logger       *zap.Logger
	services     services
	repositories repositories
}

func (s *App) Run() error {
	errgroup := new(errgroup.Group)
	errgroup.Go(func() error {
		return s.httpServer.Run()
	})
	errgroup.Go(func() error {
		return s.grpcServer.Run()
	})
	if err := errgroup.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *App) Shutdown() {
	// notify all subscribed goroutines to exit
	s.httpServer.Shutdown()
	s.grpcServer.Shutdown()
	close(s.stopCh)
	if s.shouldWaitForDone {
		<-s.doneCh
	}
	<-s.auditDoneCh
	s.opts.logger.Info("All goroutines have exited")
}

func NewApp(v *appenv.Variables) *App {
	// db
	if v.DatabaseDSN == nil {
		log.Fatal("database dsn is not set")
	}
	dbConf := db.NewDBConfig(*v.DatabaseDSN)
	d, _ := driver.NewSQLDriver(dbConf)

	// logger
	logger, err := logger.NewLogger(zap.NewAtomicLevelAt(zap.InfoLevel))
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	// channels
	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	auditDoneCh := make(chan struct{})

	// repositories
	metricRepo := repository.NewMetricRepository(
		*v.FileStoragePath,
		*v.Restore,
		time.Second*time.Duration(*v.StoreInterval),
		d,
		stopCh,
		doneCh,
	)

	// read RSA private key if provided
	var privKey *rsa.PrivateKey
	if *v.CryptoKey != "" {
		privKey, err = utils.ReadRSAPrivateKeyFromFile(*v.CryptoKey)
		if err != nil {
			log.Fatalf("failed to read RSA private key: %v", err)
		}
	}

	// services
	auditService := service.NewAuditService(*v.AuditFile, *v.AuditURL, stopCh, auditDoneCh)
	metricService := service.NewMetricService(metricRepo, []byte(*v.Key), auditService, privKey)
	baseOpts := baseServerOpts{
		vars:   v,
		logger: logger,
		services: services{
			metricsService: metricService,
		},
		repositories: repositories{
			metricRepo: metricRepo,
		},
	}

	http := NewHTTPServer(baseOpts)
	grpc := NewGRPCServer(baseOpts)

	return &App{
		httpServer:        http,
		grpcServer:        grpc,
		stopCh:            stopCh,
		doneCh:            doneCh,
		auditDoneCh:       auditDoneCh,
		shouldWaitForDone: *v.StoreInterval != 0,
		opts:              baseOpts,
	}
}
