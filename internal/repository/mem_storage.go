package repository

import (
	"errors"
	"fmt"
	"sync"
)

var MemStorage = NewMemStorage()

type MemoryStorage struct {
	namedCounter map[string]int64
	namedGauge   map[string]float64

	mutex sync.RWMutex
}

func NewMemStorage() *MemoryStorage {
	return &MemoryStorage{namedGauge: make(map[string]float64), namedCounter: make(map[string]int64)}
}

func (s *MemoryStorage) IncrementCounter(name string, number int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.namedCounter[name]; ok {
		s.namedCounter[name] = s.namedCounter[name] + number
	} else {
		s.namedCounter[name] = number
	}
}

func (s *MemoryStorage) SetGauge(name string, gauge float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.namedGauge[name] = gauge
}

var ErrNotFound = errors.New("not found")

func (s *MemoryStorage) GetGauge(name string) (float64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, ok := s.namedGauge[name]
	if !ok {
		return 0.0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return value, nil
}

func (s *MemoryStorage) GetCounter(name string) (int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, ok := s.namedCounter[name]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return value, nil
}
