package domain_model

import "net/http"

type MetricService interface {
	GetMetrics(w http.ResponseWriter, r *http.Request)
	SetMetrics(w http.ResponseWriter, r *http.Request)
}
