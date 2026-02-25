package domain_model

type MetricRepository interface {
	IncrementCounter(name string, number int64)
	SetGauge(name string, gauge float64)
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
}
