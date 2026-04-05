package handler

import (
	"github.com/gorilla/mux"
)

func NewRouter(handler *Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handler.SetMetrics)
	r.HandleFunc("/value/{metricType}/{metricName}", handler.GetMetrics)
	r.HandleFunc("/update/", handler.SetMetricsUniversal)
	r.HandleFunc("/update", handler.SetMetricsUniversal)
	r.HandleFunc("/value/", handler.GetMetricsUniversal)
	r.HandleFunc("/value", handler.GetMetricsUniversal)
	r.HandleFunc("/", handler.GetAllMetrics)
	return r
}
