package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/gorilla/mux"
)

func PostMetrics(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "invalid method, expected post", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "invalid Content-Type: expected text/plain", http.StatusBadRequest)
		return
	}

	requestComponents := mux.Vars(r)

	metricTypeStr := requestComponents["metricType"]
	metricType := MetricType(metricTypeStr)
	metricName := requestComponents["metricName"]
	metricValue := requestComponents["metricValue"]

	if _, found := validMetricTypes[metricType]; !found {
		http.Error(w, "invalid metric type: "+metricTypeStr, http.StatusBadRequest)
		return
	}

	log.Printf("Get metrics for %s %s %s", metricType, metricName, metricValue)
	// let's do a request sanity check
	if metricName == "" {
		http.Error(w, "metric name is empty", http.StatusBadRequest)
		return
	}
	if metricValue == "" {
		http.Error(w, "metric value is empty", http.StatusBadRequest)
		return
	}

	switch metricType {
	case Gauge:

		metricValueGauge, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "invalid gauge value: "+metricValue+err.Error(), http.StatusBadRequest)
			return
		}
		repository.MemStorage.SetGauge(metricName, metricValueGauge)
	case Counter:
		metricValueCounter, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "invalid counter value: "+metricValue+err.Error(), http.StatusBadRequest)
			return
		}
		repository.MemStorage.IncrementCounter(metricName, metricValueCounter)
	}
	w.WriteHeader(http.StatusOK)
}
