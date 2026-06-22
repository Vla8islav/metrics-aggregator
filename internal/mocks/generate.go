package mocks

//go:generate mockgen -destination=domain_mock.go -package=mocks github.com/Vla8islav/metrics-aggregator/internal/domain MetricRepository,MetricService
