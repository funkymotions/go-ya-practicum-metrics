// TODO: this package created in terms to separate gRPC procedures from HTTP handler
// and code under handler package needs to be reorganized.
package grpc

import (
	"context"
	"encoding/json"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/proto"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type metricService interface {
	SetMetricBulk([]byte, []byte, string) error
}

type metricGRPCHandler struct {
	proto.UnimplementedMetricsServer
	service metricService
}

func NewMetricGRPCHandler(service metricService) *metricGRPCHandler {
	return &metricGRPCHandler{
		service: service,
	}
}

func (h *metricGRPCHandler) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	protoMetrics := req.GetMetrics()
	if len(protoMetrics) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no metrics provided")
	}

	modelMetrics := utils.CastProtoMetricsToModel(protoMetrics)

	// FIXME: need to rework service.SetMetricBulk form []byte to models.Metrics
	// to reduce deserialization into JSON due to speed up
	payload, err := json.Marshal(modelMetrics)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to serialize payload")
	}

	err = h.service.SetMetricBulk(payload, nil, "")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set metrics: %v", err)
	}

	return &proto.UpdateMetricsResponse{}, nil
}
