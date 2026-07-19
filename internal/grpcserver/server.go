package grpcserver

import (
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	"github.com/Vla8islav/metrics-aggregator/internal/proto"
)

type Server struct {
	proto.MetricService

	service domain.MetricService
}

func New(service domain.MetricService) *Server {
	return &Server{
		service: service,
	}
}
