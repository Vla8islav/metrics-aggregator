package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBatchSetGauge(t *testing.T) {
	storage := getTestStorage(t)

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
	storage := getTestStorage(t)

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
	storage := getTestStorage(t)
	ctx := context.Background()

	err := storage.batchSetGauge(ctx,
		[]string{"g1", "g2"},
		[]float64{1.0},
	)

	require.Error(t, err)
}

func TestBatchSetGauge_Empty(t *testing.T) {
	storage := getTestStorage(t)
	ctx := context.Background()

	err := storage.batchSetGauge(ctx, nil, nil)
	require.NoError(t, err)
}

func TestBatchIncrementCounters(t *testing.T) {
	storage := getTestStorage(t)
	ctx := context.Background()

	err := storage.batchIncrementCounters(ctx,
		[]string{"c1", "c2", "c3"},
		[]int64{1, 2, 3},
	)
	require.NoError(t, err)

	c1, err := storage.GetCounter(ctx, "c1")
	require.NoError(t, err)
	require.Equal(t, int64(1), c1)

	c2, err := storage.GetCounter(ctx, "c2")
	require.NoError(t, err)
	require.Equal(t, int64(2), c2)

	c3, err := storage.GetCounter(ctx, "c3")
	require.NoError(t, err)
	require.Equal(t, int64(3), c3)
}

func TestBatchIncrementCounters_UpdateExisting(t *testing.T) {
	storage := getTestStorage(t)
	ctx := context.Background()

	err := storage.batchIncrementCounters(ctx,
		[]string{"c1", "c2"},
		[]int64{1, 2},
	)
	require.NoError(t, err)

	err = storage.batchIncrementCounters(ctx,
		[]string{"c1", "c2", "c3"},
		[]int64{10, 20, 30},
	)
	require.NoError(t, err)

	c1, err := storage.GetCounter(ctx, "c1")
	require.NoError(t, err)
	require.Equal(t, int64(11), c1)

	c2, err := storage.GetCounter(ctx, "c2")
	require.NoError(t, err)
	require.Equal(t, int64(22), c2)

	c3, err := storage.GetCounter(ctx, "c3")
	require.NoError(t, err)
	require.Equal(t, int64(30), c3)
}

func TestBatchIncrementCounters_LengthMismatch(t *testing.T) {
	storage := getTestStorage(t)
	ctx := context.Background()

	err := storage.batchIncrementCounters(ctx,
		[]string{"c1", "c2"},
		[]int64{1},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "incorrect number of counters")
}

func TestBatchIncrementCounters_Empty(t *testing.T) {
	storage := getTestStorage(t)
	ctx := context.Background()

	err := storage.batchIncrementCounters(ctx, nil, nil)
	require.NoError(t, err)
}
