package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/gorilla/mux"
)

func GetAllMetrics(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestComponents := mux.Vars(r)

	metricTypeStr := requestComponents["metricType"]
	metricType := MetricType(metricTypeStr)
	metricName := requestComponents["metricName"]

	if _, found := validMetricTypes[metricType]; !found {
		log.Printf("Invalid metric type: %s", metricType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Get metrics for %s %s", metricType, metricName)
	// let's do a request sanity check
	if metricName == "" {
		log.Printf("Metric name is empty %s", metricName)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metricType {
	case Gauge:

		gauge, err := repository.MemStorage.GetGauge(metricName)
		if errors.Is(err, repository.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		gaugeStr := strconv.FormatFloat(gauge, 'f', -1, 64)
		_, err = w.Write([]byte(gaugeStr))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case Counter:
		counter, err := repository.MemStorage.GetCounter(metricName)
		if errors.Is(err, repository.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		counterStr := strconv.FormatInt(counter, 10)
		_, err = w.Write([]byte(counterStr))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
