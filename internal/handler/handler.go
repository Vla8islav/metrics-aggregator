package handler

import (
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
)

type Handler struct {
	service domain.MetricService
}

func NewHandler(service domain.MetricService) *Handler {
	return &Handler{service: service}
}
