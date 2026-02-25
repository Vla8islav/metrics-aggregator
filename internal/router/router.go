package router

import (
	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/gorilla/mux"
)

func NewRouter(handler *handler.Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handler.SetMetrics)
	r.HandleFunc("/value/{metricType}/{metricName}", handler.GetMetrics)
	return r
}
