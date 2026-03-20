package service

import (
	"context"

	domain "github.com/Vla8islav/metrics-aggregator/internal/domain"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
)

type metricsService struct {
	repository domain.MetricRepository
}

func NewMetricsService(repo domain.MetricRepository) domain.MetricService {
	return metricsService{repository: repo}
}

func (m metricsService) IncrementCounter(ctx context.Context, name string, number int64) error {
	return m.repository.IncrementCounter(ctx, name, number)
}

func (m metricsService) SetGauge(ctx context.Context, name string, gauge float64) error {
	return m.repository.SetGauge(ctx, name, gauge)
}

func (m metricsService) GetGauge(ctx context.Context, name string) (float64, error) {
	return m.repository.GetGauge(ctx, name)
}

func (m metricsService) GetCounter(ctx context.Context, name string) (int64, error) {
	return m.repository.GetCounter(ctx, name)
}

func (m metricsService) GetAll(ctx context.Context) (models.MetricsExport, error) {
	return m.repository.GetAll(ctx)
}
