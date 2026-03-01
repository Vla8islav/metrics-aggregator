package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	models "github.com/Vla8islav/metrics-aggregator/internal/model"
)

type MemoryStorage struct {
	namedCounter map[string]int64
	namedGauge   map[string]float64

	mu sync.RWMutex
}

func NewMemStorage() *MemoryStorage {
	return &MemoryStorage{namedGauge: make(map[string]float64), namedCounter: make(map[string]int64)}
}

func (s *MemoryStorage) GetAll(ctx context.Context) (models.MetricsExport, error) {
	select {
	case <-ctx.Done():
		return models.MetricsExport{}, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	counters := make(map[string]int64)
	for k, v := range s.namedCounter {
		counters[k] = v
	}

	gauges := make(map[string]float64)
	for k, v := range s.namedGauge {
		gauges[k] = v
	}

	return models.MetricsExport{
		Counters: counters,
		Gauges:   gauges,
	}, nil

}

func (s *MemoryStorage) IncrementCounter(ctx context.Context, name string, number int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.namedCounter[name]; ok {
		s.namedCounter[name] = s.namedCounter[name] + number
	} else {
		s.namedCounter[name] = number
	}
	return nil
}

func (s *MemoryStorage) SetGauge(ctx context.Context, name string, gauge float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.namedGauge[name] = gauge
	return nil
}

var ErrNotFound = errors.New("not found")

func (s *MemoryStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	select {
	case <-ctx.Done():
		return 0.0, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.namedGauge[name]
	if !ok {
		return 0.0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return value, nil
}

func (s *MemoryStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	select {
	case <-ctx.Done():
		return 0.0, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.namedCounter[name]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return value, nil
}
