package domain

import (
	"context"
)

type MetricServer interface {
	SetMetrics(ctx context.Context) error
}
