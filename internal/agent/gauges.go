package agent

import (
	"math/rand"
	"runtime"
	"sync"
)

type Stats struct {
	ms       runtime.MemStats
	gauges   map[string]float64
	counters map[string]int64

	PollCount   int64
	RandomValue float64

	mu sync.RWMutex
}

func NewStats() *Stats {
	s := Stats{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
		ms:       runtime.MemStats{},
	}
	return &s
}

func (s *Stats) GetGauges() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gauges
}

func (s *Stats) GetCounters() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counters
}

func (s *Stats) Update() error {
	s.readMemStats()
	s.incrementPollCount()
	s.updateRandomValue()

	s.mu.Lock()
	defer s.mu.Unlock()
	// maybe use reflect here
	s.gauges["Alloc"] = float64(s.ms.Alloc)
	s.gauges["BuckHashSys"] = float64(s.ms.BuckHashSys)
	s.gauges["Frees"] = float64(s.ms.Frees)
	s.gauges["GCCPUFraction"] = s.ms.GCCPUFraction
	s.gauges["GCSys"] = float64(s.ms.GCSys)
	s.gauges["HeapAlloc"] = float64(s.ms.HeapAlloc)
	s.gauges["HeapIdle"] = float64(s.ms.HeapIdle)
	s.gauges["HeapInuse"] = float64(s.ms.HeapInuse)
	s.gauges["HeapObjects"] = float64(s.ms.HeapObjects)
	s.gauges["HeapReleased"] = float64(s.ms.HeapReleased)
	s.gauges["HeapSys"] = float64(s.ms.HeapSys)
	s.gauges["LastGC"] = float64(s.ms.LastGC)
	s.gauges["Lookups"] = float64(s.ms.Lookups)
	s.gauges["MCacheInuse"] = float64(s.ms.MCacheInuse)
	s.gauges["MCacheSys"] = float64(s.ms.MCacheSys)
	s.gauges["MSpanInuse"] = float64(s.ms.MSpanInuse)
	s.gauges["MSpanSys"] = float64(s.ms.MSpanSys)
	s.gauges["Mallocs"] = float64(s.ms.Mallocs)
	s.gauges["NextGC"] = float64(s.ms.NextGC)
	s.gauges["NumForcedGC"] = float64(s.ms.NumForcedGC)
	s.gauges["NumGC"] = float64(s.ms.NumGC)
	s.gauges["OtherSys"] = float64(s.ms.OtherSys)
	s.gauges["PauseTotalNs"] = float64(s.ms.PauseTotalNs)
	s.gauges["StackInuse"] = float64(s.ms.StackInuse)
	s.gauges["StackSys"] = float64(s.ms.StackSys)
	s.gauges["Sys"] = float64(s.ms.Sys)
	s.gauges["TotalAlloc"] = float64(s.ms.TotalAlloc)

	// additional random value
	s.gauges["RandomValue"] = s.RandomValue
	// counter
	s.counters["PollCount"] = s.PollCount

	return nil
}

func (s *Stats) updateRandomValue() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RandomValue = rand.Float64()
}

func (s *Stats) incrementPollCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PollCount++
}

func (s *Stats) readMemStats() {
	s.mu.Lock()
	defer s.mu.Unlock()
	runtime.ReadMemStats(&s.ms)
}
