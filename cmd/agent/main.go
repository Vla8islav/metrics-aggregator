package main

import (
	"math/rand"
	"runtime"
	"sync"
)

type Metrics struct {
	Ms          runtime.MemStats
	PollCount   int64
	RandomValue float64

	mu sync.Mutex
}

func (m *Metrics) updateRandomValue() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RandomValue = rand.Float64()
}

func (m *Metrics) incrementPollCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PollCount++
}

func (m *Metrics) readMemStats() {
	m.mu.Lock()
	defer m.mu.Unlock()
	runtime.ReadMemStats(&m.Ms)
}

func main() {
	// Разработайте агент (HTTP-клиент) для сбора рантайм-метрик и их последующей отправки на сервер по протоколу HTTP.
	// Собираем
	metrics := Metrics{}

	metrics.updateRandomValue()
	metrics.readMemStats()
	metrics.incrementPollCount()

	// Отправляем на сервер

}
