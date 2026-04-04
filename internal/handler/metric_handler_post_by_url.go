package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/gorilla/mux"
)

func (h *Handler) SetMetrics(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Println("Only POST method is allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestComponents := mux.Vars(r)

	metricTypeStr := requestComponents["metricType"]
	metricType := models.MetricType(metricTypeStr)
	metricName := requestComponents["metricName"]
	metricValue := requestComponents["metricValue"]

	if _, found := models.ValidMetricTypes[metricType]; !found {
		writeBadRequest(w, "invalid metric type: "+metricTypeStr)
		return
	}

	log.Printf("Post metrics for %s %s %s", metricType, metricName, metricValue)
	// let's do a request sanity check
	if metricName == "" {
		writeBadRequest(w, "metric name is empty")
		return
	}
	if metricValue == "" {
		writeBadRequest(w, "metric value is empty")
		return
	}

	switch metricType {
	case models.Gauge:

		metricValueGauge, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			writeBadRequest(w, "invalid gauge value: "+metricValue+err.Error())
			return
		}
		err = h.service.SetGauge(r.Context(), metricName, metricValueGauge)
		if err != nil {
			log.Println("error when setting gauge: ", err, metricName, metricValue)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case models.Counter:
		metricValueCounter, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			writeBadRequest(w, "invalid counter value: "+metricValue+err.Error())
			return
		}
		err = h.service.IncrementCounter(r.Context(), metricName, metricValueCounter)
		if err != nil {
			log.Println("error when setting gauge: ", err, metricName, metricValue)
			w.WriteHeader(http.StatusInternalServerError)
			return

		}
	}
	w.WriteHeader(http.StatusOK)
}
