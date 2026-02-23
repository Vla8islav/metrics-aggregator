package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/gorilla/mux"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

var validMetricTypes = map[MetricType]struct{}{Gauge: {}, Counter: {}}
var memStorage = repository.NewMemStorage()

func GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	requestComponents := mux.Vars(r)

	metricTypeStr := requestComponents["metricType"]
	metricType := MetricType(metricTypeStr)
	metricName := requestComponents["metricName"]
	metricValue := requestComponents["metricValue"]

	if _, found := validMetricTypes[metricType]; !found {
		log.Printf("Invalid metric type: %s", metricType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Get metrics for %s %s %s", metricType, metricName, metricValue)
	// let's do a request sanity check
	if metricName == "" {
		log.Printf("Metric name is empty %s", metricName)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if metricValue == "" {
		log.Printf("Metric has an empty value '%s'", metricValue)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch metricType {
	case Gauge:

		metricValueGauge, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			log.Printf("Error parsing metric value: %s", metricValue)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		memStorage.SetGauge(metricValueGauge)
	case Counter:
		metricValueCounter, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			log.Printf("Error parsing metric value: %s", metricValue)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		memStorage.IncrementCounter(metricValueCounter)
	}
	w.WriteHeader(http.StatusOK)

}
