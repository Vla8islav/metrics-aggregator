package models

// MetricType identifies metric value type
type MetricType string

const (
	// Gauge represents a metric that stores the latest floating-point value
	Gauge MetricType = "gauge"

	// Counter represents a metric that accumulates integer deltas
	Counter MetricType = "counter"
)

// ValidMetricTypes contains all metric types accepted by the application
var ValidMetricTypes = map[MetricType]struct{}{Gauge: {}, Counter: {}}

// Metrics represents a metric update or value exchanged through the API.
type Metrics struct {
	ID    string     `json:"id"`              // metric name
	MType MetricType `json:"type"`            // metric type: gauge or counter
	Delta *int64     `json:"delta,omitempty"` // counter value or delta
	Value *float64   `json:"value,omitempty"` // gauge value
}

// MetricsExport groups metric values by type for rendering or export
type MetricsExport struct {
	Counters map[string]int64
	Gauges   map[string]float64
}
