package domain

import (
	"context"

	models "github.com/Vla8islav/metrics-aggregator/internal/model"
)

type MetricService interface {
	IncrementCounter(ctx context.Context, name string, number int64) error
	SetGauge(ctx context.Context, name string, gauge float64) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (models.MetricsExport, error)
}
