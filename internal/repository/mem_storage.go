package repository

import (
	"sync"
)

type MemStorage struct {
	counter int64
	gauge   float64

	mutex sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{counter: 0, gauge: 0}
}

func (s *MemStorage) IncrementCounter(number int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.counter += number
}

func (s *MemStorage) SetGauge(gauge float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.gauge = gauge
}

func (s *MemStorage) GetGauge() float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.gauge
}

func (s *MemStorage) GetCounter() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.counter
}
