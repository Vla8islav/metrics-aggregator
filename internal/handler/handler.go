package handler

import (
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	"go.uber.org/zap"
)

// Handler groups HTTP handlers with their service dependency and logger
type Handler struct {
	service domain.MetricService
	logger  *zap.Logger
}

// NewHandler creates a Handler with the provided metric service and logger
func NewHandler(service domain.MetricService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}
