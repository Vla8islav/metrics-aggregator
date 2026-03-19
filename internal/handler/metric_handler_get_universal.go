package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func (h *Handler) GetMetricsUniversal(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Println("Only POST method is allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		writeBadRequest(w, "only application/json content type is supported")
		return
	}

	requestBody, err := io.ReadAll(r.Body)
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
		val, err := h.service.GetGauge(r.Context(), requestBodySerialised.ID)
		if err != nil {
			log.Println("error when getting the gauge: ", requestBodySerialised.ID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		responseBodySerialized := Metrics{
			ID:    requestBodySerialised.ID,
			MType: string(Gauge),
			Value: &val,
		}
		w.Header().Set("Content-Type", "application/json")
		responseBodySerializedJson, err := json.Marshal(responseBodySerialized)
		if err != nil {
			log.Println("error when marshalling response body gauge: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(responseBodySerializedJson)
		w.WriteHeader(http.StatusOK)
		return

	case Counter:
		val, err := h.service.GetCounter(r.Context(), requestBodySerialised.ID)
		if err != nil {
			log.Println("error when getting the counter: ", requestBodySerialised.ID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		responseBodySerialized := Metrics{
			ID:    requestBodySerialised.ID,
			MType: string(Counter),
			Delta: &val,
		}
		w.Header().Set("Content-Type", "application/json")
		responseBodySerializedJson, err := json.Marshal(responseBodySerialized)
		if err != nil {
			log.Println("error when marshalling response body counter: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(responseBodySerializedJson)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusOK)
}
