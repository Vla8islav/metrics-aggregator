package domain

import (
	"context"

	models "github.com/Vla8islav/metrics-aggregator/internal/model"
)

// MetricService defines business operations for collecting and reading metrics
type MetricService interface {
	// Ping checks whether the service dependencies are available
	Ping(ctx context.Context) error
	// IncrementCounter adds number to the named counter
	IncrementCounter(ctx context.Context, name string, number int64) error
	// SetGauge stores the current value of the named gauge
	SetGauge(ctx context.Context, name string, gauge float64) error
	// GetGauge returns the current value of the named gauge
	GetGauge(ctx context.Context, name string) (float64, error)
	// GetCounter returns the current value of the named counter
	GetCounter(ctx context.Context, name string) (int64, error)
	// GetAll exports all the currently saved metrics
	GetAll(ctx context.Context) (models.MetricsExport, error)
	// UpdateMetrics batch updates metrics
	UpdateMetrics(ctx context.Context, input []models.Metrics) error
}
