package repository

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorageIncrementCounter(t *testing.T) {
	s := NewMemStorage()

	s.IncrementCounter(5)
	s.IncrementCounter(3)

	assert.EqualValues(t, 8, s.GetCounter())

}

func TestMemStorageSetGauge(t *testing.T) {
	s := NewMemStorage()
	delta := 0.000001
	s.SetGauge(42.5)

	assert.InDelta(t, 42.5, s.GetGauge(), delta)
	s.SetGauge(-12.5)
	assert.InDelta(t, -12.5, s.GetGauge(), delta)
}

func TestMemStorageConcurrentAccessCounter(t *testing.T) {
	s := NewMemStorage()
	var wg sync.WaitGroup
	const workers = 100
	incrementsPerWorker := 10
	incrementStep := int64(1)
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerWorker; j++ {
				s.IncrementCounter(incrementStep)
			}
		}()
	}
	wg.Wait()

	expectedResult := int64(workers*incrementsPerWorker) * incrementStep

	assert.Equal(t, expectedResult, s.GetCounter())

}

func TestMemStorageConcurrentAccessGauge(t *testing.T) {
	s := NewMemStorage()
	var wg sync.WaitGroup
	const workers = 100
	wg.Add(workers)
	values := make(map[float64]struct{}, workers)

	var mu sync.Mutex

	for i := 0; i < workers; i++ {
		v := float64(i) + 0.1

		mu.Lock()
		values[v] = struct{}{}
		mu.Unlock()

		go func(val float64) {
			defer wg.Done()
			s.SetGauge(val)
		}(v)
	}

	wg.Wait()

	finalValue := s.GetGauge()

	_, got := values[finalValue]

	assert.True(t, got)
}
