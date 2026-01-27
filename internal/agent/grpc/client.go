package grpc

import (
	"context"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AgentGRPCClient struct {
	conn   any
	client proto.MetricsClient
}

func NewAgentGRPCClient(conn *grpc.ClientConn) *AgentGRPCClient {
	m := proto.NewMetricsClient(conn)

	return &AgentGRPCClient{
		client: m,
	}
}

func (c *AgentGRPCClient) SendMetrics(ctx context.Context, metrics []*models.Metrics, agentIP string) error {
	md := metadata.Pairs("x-real-ip", agentIP)
	ctx = metadata.NewOutgoingContext(ctx, md)
	var protoMetrics []*proto.Metric
	for _, m := range metrics {
		metric := &proto.Metric{
			Id: m.ID,
		}
		switch m.MType {
		case models.Gauge:
			metric.Type = proto.Metric_GAUGE
			if m.Value != nil {
				metric.Value = *m.Value
			}
		case models.Counter:
			metric.Type = proto.Metric_COUNTER
			if m.Delta != nil {
				metric.Delta = *m.Delta
			}
		}
		protoMetrics = append(protoMetrics, metric)
	}

	_, err := c.client.UpdateMetrics(ctx, &proto.UpdateMetricsRequest{
		Metrics: protoMetrics,
	})

	return err
}
