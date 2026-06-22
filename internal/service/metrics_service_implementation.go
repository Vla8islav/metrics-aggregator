package service

import (
	"context"

	domain "github.com/Vla8islav/metrics-aggregator/internal/domain"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
)

// metricsService implements MetricService by delegating metric operations to a repository
type metricsService struct {
	repository domain.MetricRepository
}

// NewMetricsService creates a MetricService backed by repo
func NewMetricsService(repo domain.MetricRepository) domain.MetricService {
	return metricsService{repository: repo}
}

// Ping checks whether the service dependencies are available
func (m metricsService) Ping(ctx context.Context) error {
	return m.repository.Ping(ctx)
}

// IncrementCounter adds number to the named counter metric
func (m metricsService) IncrementCounter(ctx context.Context, name string, number int64) error {
	return m.repository.IncrementCounter(ctx, name, number)
}

// SetGauge stores the current value of the named gauge metric
func (m metricsService) SetGauge(ctx context.Context, name string, gauge float64) error {
	return m.repository.SetGauge(ctx, name, gauge)
}

// GetGauge returns the current value of the named gauge metric
func (m metricsService) GetGauge(ctx context.Context, name string) (float64, error) {
	return m.repository.GetGauge(ctx, name)
}

// GetCounter returns the current value of the named counter
func (m metricsService) GetCounter(ctx context.Context, name string) (int64, error) {
	return m.repository.GetCounter(ctx, name)
}

// GetAll returns all currently stored metrics grouped for export
func (m metricsService) GetAll(ctx context.Context) (models.MetricsExport, error) {
	return m.repository.GetAll(ctx)
}

// UpdateMetrics applies a batch of metric updates
func (m metricsService) UpdateMetrics(ctx context.Context, input []models.Metrics) error {
	return m.repository.UpdateMetrics(ctx, input)
}
