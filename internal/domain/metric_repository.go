package domain

import (
	"context"

	"github.com/Vla8islav/metrics-aggregator/internal/model"
)

type MetricRepository interface {
	IncrementCounter(ctx context.Context, name string, number int64) error
	SetGauge(ctx context.Context, name string, gauge float64) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (models.MetricsExport, error)
	SaveState(ctx context.Context) error
	LoadState(ctx context.Context) error
}
