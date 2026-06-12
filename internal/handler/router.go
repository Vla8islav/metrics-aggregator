package handler

import (
	"github.com/gorilla/mux"
)

// NewRouter creates a router with all metrics API routes registered
//
// The returned router exposes endpoints for updating metrics, reading metrics,
// checking database connectivity, batch updates, and rendering all metrics.
func NewRouter(handler *Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handler.SetMetrics)
	r.HandleFunc("/value/{metricType}/{metricName}", handler.GetMetrics)
	r.HandleFunc("/update/", handler.SetMetricsUniversal)
	r.HandleFunc("/update", handler.SetMetricsUniversal)
	r.HandleFunc("/value/", handler.GetMetricsUniversal)
	r.HandleFunc("/value", handler.GetMetricsUniversal)
	r.HandleFunc("/ping/", handler.DBPing)
	r.HandleFunc("/ping", handler.DBPing)
	r.HandleFunc("/updates/", handler.UpdateBatchMetrics)
	r.HandleFunc("/updates", handler.UpdateBatchMetrics)
	r.HandleFunc("/", handler.GetAllMetrics)
	return r
}
