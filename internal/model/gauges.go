// Package models data models for the metric aggregation server
package models

import (
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Stats stores runtime, memory, CPU, gauge, and counter collected by the agent
type Stats struct {
	ms             runtime.MemStats
	vmStat         *mem.VirtualMemoryStat
	cpuUtilisation []float64

	gauges   map[string]float64
	counters map[string]int64

	PollCount   int64
	RandomValue float64

	mu sync.RWMutex
}

// NewStats creates an initialized Stats value ready for metric collection
func NewStats() *Stats {
	s := Stats{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
		ms:       runtime.MemStats{},
	}
	return &s
}

// GetGauges returns a copy of the currently collected gauge metrics
func (s *Stats) GetGauges() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		out[k] = v
	}
	return out
}

// GetCounters returns a copy of the currently collected counter
func (s *Stats) GetCounters() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		out[k] = v
	}
	return out
}

// Update refreshes runtime, memory, CPU, gauge, and counter
func (s *Stats) Update() error {
	s.readMemStats()
	err := s.readAdditionalMemStats()
	if err != nil {
		return err
	}
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

	// additional info
	s.gauges["TotalMemory"] = float64(s.vmStat.Total)
	s.gauges["FreeMemory"] = float64(s.vmStat.Free)
	s.gauges["CPUutilization1"] = s.cpuUtilisation[0]

	return nil
}

// updateRandomValue stores a new random gauge value
func (s *Stats) updateRandomValue() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RandomValue = rand.Float64()
}

// incrementPollCount increments the number of completed metric polls
func (s *Stats) incrementPollCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PollCount++
}

// readMemStats reads Go runtime memory statistics
func (s *Stats) readMemStats() {
	s.mu.Lock()
	defer s.mu.Unlock()
	runtime.ReadMemStats(&s.ms)
}

// readAdditionalMemStats reads host memory and CPU utilization statistics
func (s *Stats) readAdditionalMemStats() error {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	// this thing is slow
	cpUtilisation, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.vmStat = vm
	s.cpuUtilisation = cpUtilisation

	return nil
}
