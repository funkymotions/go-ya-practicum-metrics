package server

import (
	"log"
	"net"

	grpc_handler "github.com/funkymotions/go-ya-practicum-metrics/internal/handler/grpc"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/logger"
	grpc_middleware "github.com/funkymotions/go-ya-practicum-metrics/internal/middleware/grpc"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type grpcServer struct {
	grpc     *grpc.Server
	logger   *zap.Logger
	endpoint string
}

func NewGRPCServer(
	opts baseServerOpts,
) *grpcServer {
	metricServer := grpc_handler.NewMetricGRPCHandler(opts.services.metricsService)
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.UnaryCIDRInterceptor(*opts.vars.TrustedSubnet)),
	)

	proto.RegisterMetricsServer(srv, metricServer)
	lo, err := logger.NewLogger(zap.NewAtomicLevelAt(zap.InfoLevel))
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	return &grpcServer{
		logger:   lo,
		grpc:     srv,
		endpoint: *opts.vars.EndpointGRPC,
	}
}

func (s *grpcServer) Run() error {
	s.logger.Info("Starting gRPC server", zap.String("addr", s.endpoint))
	l, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		return err
	}
	if err := s.grpc.Serve(l); err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) Shutdown() {
	s.logger.Warn("Shutting down gRPC server")
	s.grpc.GracefulStop()
	s.logger.Info("gRPC server has shut down")
}
