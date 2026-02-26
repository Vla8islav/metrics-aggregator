package handler

import (
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
)

type Handler struct {
	repo domain.MetricRepository
}

func NewHandler(repository domain.MetricRepository) *Handler {
	return &Handler{repo: repository}
}
