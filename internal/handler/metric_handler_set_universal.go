package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/model"
)

func (h *Handler) SetMetricsUniversal(w http.ResponseWriter, r *http.Request) {

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

		if requestBodySerialised.Value == nil {
			h.writeBadRequest(w, "gauge value cannot be nil")
			return
		}

		err = h.service.SetGauge(r.Context(), requestBodySerialised.ID, *requestBodySerialised.Value)
		if err != nil {
			log.Println("error when setting gauge: ", err, requestBodySerialised.ID, *requestBodySerialised.Value)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case models.Counter:
		if requestBodySerialised.Delta == nil {
			h.writeBadRequest(w, "gauge value cannot be nil")
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
