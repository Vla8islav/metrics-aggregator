package handler

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

var validMetricTypes = map[MetricType]struct{}{Gauge: {}, Counter: {}}
