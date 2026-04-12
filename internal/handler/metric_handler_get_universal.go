package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
)

func (h *Handler) GetMetricsUniversal(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		h.writeMethodNotAllowed(w, "only POST method is allowed")
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		h.writeBadRequest(w, "only application/json content type is supported")
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeBadRequest(w, "failed to read request body: "+err.Error())
		return
	}

	var requestBodySerialised models.Metrics
	err = json.Unmarshal(requestBody, &requestBodySerialised)
	if err != nil {
		h.writeBadRequest(w, err.Error())
		return
	}

	metricType := models.MetricType(requestBodySerialised.MType)
	if _, found := models.ValidMetricTypes[metricType]; !found {
		h.writeBadRequest(w, "invalid metric type: "+string(metricType))
		return
	}
	if requestBodySerialised.ID == "" {
		h.writeBadRequest(w, "metric ID cannot be empty")
		return
	}

	switch metricType {
	case models.Gauge:
		val, err := h.service.GetGauge(r.Context(), requestBodySerialised.ID)
		if errors.Is(err, repository.ErrNotFound) {
			log.Println("gauge with this name wasn't found: ", requestBodySerialised.ID, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Println("error when getting the gauge: ", requestBodySerialised.ID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		responseBodySerialized := models.Metrics{
			ID:    requestBodySerialised.ID,
			MType: models.Gauge,
			Value: &val,
		}
		w.Header().Set("Content-Type", "application/json")
		responseBodySerializedJSON, err := json.Marshal(responseBodySerialized)
		if err != nil {
			log.Println("error when marshalling response body gauge: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(responseBodySerializedJSON)
		w.WriteHeader(http.StatusOK)
		return

	case models.Counter:
		val, err := h.service.GetCounter(r.Context(), requestBodySerialised.ID)
		if errors.Is(err, repository.ErrNotFound) {
			log.Println("counter with this name wasn't found: ", requestBodySerialised.ID, err)
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("error when getting the counter: ", requestBodySerialised.ID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		responseBodySerialized := models.Metrics{
			ID:    requestBodySerialised.ID,
			MType: models.Counter,
			Delta: &val,
		}
		w.Header().Set("Content-Type", "application/json")
		responseBodySerializedJSON, err := json.Marshal(responseBodySerialized)
		if err != nil {
			log.Println("error when marshalling response body counter: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(responseBodySerializedJSON)
		if err != nil {
			log.Println("error when writing response body counter: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusOK)
}
