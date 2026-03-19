package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func (h *Handler) SetMetricsUniversal(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Println("Only POST method is allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		writeBadRequest(w, "only application/json content type is supported")
		return
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeBadRequest(w, "failed to read request body: "+err.Error())
		return
	}

	var requestBodySerialised Metrics
	err = json.Unmarshal(requestBody, &requestBodySerialised)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	metricType := MetricType(requestBodySerialised.MType)
	if _, found := validMetricTypes[metricType]; !found {
		writeBadRequest(w, "invalid metric type: "+string(metricType))
		return
	}

	if requestBodySerialised.ID == "" {
		writeBadRequest(w, "metric ID cannot be empty")
		return
	}

	switch metricType {
	case Gauge:

		if requestBodySerialised.Value == nil {
			writeBadRequest(w, "gauge value cannot be nil")
			return
		}

		err = h.service.SetGauge(r.Context(), requestBodySerialised.ID, *requestBodySerialised.Value)
		if err != nil {
			log.Println("error when setting gauge: ", err, requestBodySerialised.ID, *requestBodySerialised.Value)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case Counter:
		if requestBodySerialised.Delta == nil {
			writeBadRequest(w, "gauge value cannot be nil")
			return
		}
		err = h.service.IncrementCounter(r.Context(), requestBodySerialised.ID, *requestBodySerialised.Delta)
		if err != nil {
			log.Println("error when setting gauge: ", err, requestBodySerialised.ID, *requestBodySerialised.Delta)
			w.WriteHeader(http.StatusInternalServerError)
			return

		}
	}
	w.WriteHeader(http.StatusOK)
}
