package models

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

var ValidMetricTypes = map[MetricType]struct{}{Gauge: {}, Counter: {}}

type Metrics struct {
	ID    string     `json:"id"`              // имя метрики
	MType MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64     `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type MetricsExport struct {
	Counters map[string]int64
	Gauges   map[string]float64
}
