package utils

import (
	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/proto"
)

func CastProtoMetricsToModel(in []*proto.Metric) []*models.Metrics {
	out := make([]*models.Metrics, 0, len(in))
	for _, m := range in {
		metricModel := models.Metrics{
			ID: m.GetId(),
		}
		switch m.GetType() {
		case proto.Metric_GAUGE:
			value := m.GetValue()
			metricModel.Value = &value
			metricModel.MType = models.Gauge
		case proto.Metric_COUNTER:
			delta := m.GetDelta()
			metricModel.Delta = &delta
			metricModel.MType = models.Counter
		}
		out = append(out, &metricModel)
	}

	return out
}
