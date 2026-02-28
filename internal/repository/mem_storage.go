package repository

import (
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

func (s *MemoryStorage) GetAll() (models.MetricsExport, error) {
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

func (s *MemoryStorage) IncrementCounter(name string, number int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.namedCounter[name]; ok {
		s.namedCounter[name] = s.namedCounter[name] + number
	} else {
		s.namedCounter[name] = number
	}
}

func (s *MemoryStorage) SetGauge(name string, gauge float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.namedGauge[name] = gauge
}

var ErrNotFound = errors.New("not found")

func (s *MemoryStorage) GetGauge(name string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.namedGauge[name]
	if !ok {
		return 0.0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return value, nil
}

func (s *MemoryStorage) GetCounter(name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.namedCounter[name]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return value, nil
}
