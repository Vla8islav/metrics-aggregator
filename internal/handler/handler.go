package handler

import (
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	service domain.MetricService
	logger  *zap.Logger
}

func NewHandler(service domain.MetricService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}
