package domain

import (
	"context"

	models "github.com/Vla8islav/metrics-aggregator/internal/model"
)

// MetricRepository defines persistent storage operations for application metrics.
type MetricRepository interface {
	// Ping checks whether the underlying storage is available
	Ping(ctx context.Context) error

	// UpdateMetrics stores multiple metric updates in one operation
	UpdateMetrics(ctx context.Context, input []models.Metrics) error

	// IncrementCounter adds number to the named counter
	IncrementCounter(ctx context.Context, name string, number int64) error
	// SetGauge stores the current value of the named gauge metric
	SetGauge(ctx context.Context, name string, gauge float64) error

	// GetCounter returns the current value of the named counter
	GetCounter(ctx context.Context, name string) (int64, error)
	// GetGauge returns the current value of the named gauge metric
	GetGauge(ctx context.Context, name string) (float64, error)
	// GetAll returns all stored metrics grouped for export
	GetAll(ctx context.Context) (models.MetricsExport, error)

	// LoadState restores previously persisted metric values into the repository
	// redundant for anything by in-file
	LoadState(ctx context.Context) error
}
