package main

import (
	"sync"
)

type MemStorage struct {
	counter int64
	gauge   float64

	mutex sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{counter: 0, gauge: 0}
}

func (s *MemStorage) IncrementCounter(number int64) {
	s.mutex.Lock()
	s.counter += number
	s.mutex.Unlock()
}

func (s *MemStorage) SetGauge(gauge float64) {
	s.mutex.Lock()
	s.gauge = gauge
	s.mutex.Unlock()
}
