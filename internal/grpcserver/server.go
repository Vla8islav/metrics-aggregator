// Package grpcserver grpc server that accepts metrics from clients
package grpcserver

import (
	"context"
	"fmt"

	"github.com/Vla8islav/metrics-aggregator/internal/audit"
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/Vla8islav/metrics-aggregator/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ proto.MetricsServer = (*Server)(nil)

type Server struct {
	proto.UnimplementedMetricsServer

	service domain.MetricService
}

func NewGRPCServer(service domain.MetricService) *Server {
	return &Server{
		service: service,
	}
}

func (s *Server) UpdateMetrics(
	ctx context.Context,
	req *proto.UpdateMetricsRequest,
) (*proto.UpdateMetricsResponse, error) {
	audit.SetOperation(ctx, "UpdateMetrics")

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	metrics := make([]models.Metrics, 0, len(req.GetMetrics()))
	for i, metric := range req.GetMetrics() {
		converted, err2 := metricFromProto(metric)
		if err2 != nil {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"invalid metric at index %d: %v",
				i,
				err2,
			)
		}

		audit.AddMetric(ctx, converted.ID)
		metrics = append(metrics, converted)
	}

	if err2 := s.service.UpdateMetrics(ctx, metrics); err2 != nil {
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}

	return &proto.UpdateMetricsResponse{}, nil
}

func metricFromProto(metric *proto.Metric) (models.Metrics, error) {
	if metric == nil {
		return models.Metrics{}, fmt.Errorf("metric is required")
	}

	if metric.GetId() == "" {
		return models.Metrics{}, fmt.Errorf("metric id is required")
	}

	switch metric.GetType() {
	case proto.Metric_GAUGE:
		value := metric.GetValue()

		return models.Metrics{
			ID:    metric.GetId(),
			MType: models.Gauge,
			Value: &value,
		}, nil

	case proto.Metric_COUNTER:
		delta := metric.GetDelta()

		return models.Metrics{
			ID:    metric.GetId(),
			MType: models.Counter,
			Delta: &delta,
		}, nil

	default:
		return models.Metrics{}, fmt.Errorf(
			"unsupported metric type %q",
			metric.GetType().String(),
		)
	}
}
