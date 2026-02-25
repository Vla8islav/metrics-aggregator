package handler

import (
	"github.com/Vla8islav/metrics-aggregator/internal/domain_model"
)

type Handler struct {
	repo domain_model.MetricRepository
}

func NewHandler(repository domain_model.MetricRepository) *Handler {
	return &Handler{repo: repository}
}
