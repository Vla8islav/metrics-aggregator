package handler

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

var validMetricTypes = map[MetricType]struct{}{Gauge: {}, Counter: {}}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
