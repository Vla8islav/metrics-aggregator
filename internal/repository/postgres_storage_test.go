package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
)

func TestBatchSetGauge(t *testing.T) {
	storage := newTestPostgresStorage(t)
	ctx := context.Background()

	err := storage.batchSetGauge(ctx,
		[]string{"g1", "g2", "g3"},
		[]float64{1.5, 2.5, 3.5},
	)
	require.NoError(t, err)

	g1, err := storage.GetGauge(ctx, "g1")
	require.NoError(t, err)
	require.Equal(t, 1.5, g1)

	g2, err := storage.GetGauge(ctx, "g2")
	require.NoError(t, err)
	require.Equal(t, 2.5, g2)

	g3, err := storage.GetGauge(ctx, "g3")
	require.NoError(t, err)
	require.Equal(t, 3.5, g3)
}

func TestBatchSetGauge_UpdateExisting(t *testing.T) {
	storage := newTestPostgresStorage(t)
	ctx := context.Background()

	err := storage.batchSetGauge(ctx,
		[]string{"g1", "g2"},
		[]float64{1.0, 2.0},
	)
	require.NoError(t, err)

	err = storage.batchSetGauge(ctx,
		[]string{"g1"},
		[]float64{10.0},
	)
	require.NoError(t, err)

	g1, err := storage.GetGauge(ctx, "g1")
	require.NoError(t, err)
	require.Equal(t, 10.0, g1)
}

func TestBatchSetGauge_LengthMismatch(t *testing.T) {
	storage := newTestPostgresStorage(t)
	ctx := context.Background()

	err := storage.batchSetGauge(ctx,
		[]string{"g1", "g2"},
		[]float64{1.0},
	)

	require.Error(t, err)
}

func TestBatchSetGauge_Empty(t *testing.T) {
	storage := newTestPostgresStorage(t)
	ctx := context.Background()

	err := storage.batchSetGauge(ctx, nil, nil)
	require.NoError(t, err)
}

func newTestPostgresStorage(t *testing.T) *PostgresStorage {
	t.Helper()

	cfg := config.ReadFlags([]string{})

	storage, err := NewPostgresStorage(cfg, "../../migrations")
	require.NoError(t, err)

	return storage
}
