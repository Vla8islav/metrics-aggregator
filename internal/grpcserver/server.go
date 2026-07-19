package grpcserver

import (
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	"github.com/Vla8islav/metrics-aggregator/internal/proto"
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
