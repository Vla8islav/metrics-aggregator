package repository

import (
	"context"
	"math/rand"
	"sync"
	"testing"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultCounterName = "count"
const defaultGaugeName = "gauge"

func TestMemStorageIncrementCounter(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	cfg, _ := config.ReadFlagsServer(nil)
	s := NewMemStorage(cfg)

	err := s.IncrementCounter(ctx, defaultCounterName, 5)
	require.NoError(t, err)
	counter, err := s.GetCounter(ctx, defaultCounterName)
	require.NoError(t, err)
	assert.EqualValues(t, 5, counter)

	err = s.IncrementCounter(ctx, defaultCounterName, 3)
	require.NoError(t, err)
	counter, err = s.GetCounter(ctx, defaultCounterName)
	require.NoError(t, err)
	assert.EqualValues(t, 8, counter)

}

func TestMemStorageSetGauge(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	cfg, _ := config.ReadFlagsServer(nil)
	s := NewMemStorage(cfg)
	delta := 0.000001

	err := s.SetGauge(ctx, defaultGaugeName, 42.5)
	require.NoError(t, err)
	gauge, err := s.GetGauge(ctx, defaultGaugeName)
	require.NoError(t, err)
	assert.InDelta(t, 42.5, gauge, delta)

	err = s.SetGauge(ctx, defaultGaugeName, -12.5)
	require.NoError(t, err)
	gauge, err = s.GetGauge(ctx, defaultGaugeName)
	require.NoError(t, err)
	assert.InDelta(t, -12.5, gauge, delta)
}

func TestMemStorageConcurrentAccessCounter(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	cfg, _ := config.ReadFlagsServer(nil)
	s := NewMemStorage(cfg)
	var wg sync.WaitGroup
	const workers = 100
	incrementsPerWorker := 10
	incrementStep := int64(1)
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerWorker; j++ {
				err := s.IncrementCounter(ctx, defaultCounterName, incrementStep)
				require.NoError(t, err)
			}
		}()
	}
	wg.Wait()

	expectedResult := int64(workers*incrementsPerWorker) * incrementStep

	value, err := s.GetCounter(ctx, defaultCounterName)
	require.NoError(t, err)
	assert.Equal(t, expectedResult, value)

}

func TestMemStorageConcurrentAccessGauge(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	cfg, _ := config.ReadFlagsServer(nil)
	s := NewMemStorage(cfg)
	var wg sync.WaitGroup
	const workers = 100
	wg.Add(workers)
	values := make(map[float64]struct{}, workers)

	var mu sync.Mutex

	for i := 0; i < workers; i++ {
		v := float64(i) + 0.1 + rand.Float64()

		mu.Lock()
		values[v] = struct{}{}
		mu.Unlock()

		go func(val float64) {
			defer wg.Done()
			err := s.SetGauge(ctx, defaultGaugeName, val)
			require.NoError(t, err)
		}(v)
	}

	wg.Wait()

	finalValue, err := s.GetGauge(ctx, defaultGaugeName)
	require.NoError(t, err)

	_, got := values[finalValue]

	assert.True(t, got)
}
