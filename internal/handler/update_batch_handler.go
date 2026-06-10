package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/audit"
	"github.com/Vla8islav/metrics-aggregator/internal/model"
)

func (h *Handler) UpdateBatchMetrics(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "UpdateBatchMetrics")

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

	var requestBodySerialised []models.Metrics
	err = json.Unmarshal(requestBody, &requestBodySerialised)
	if err != nil {
		h.writeBadRequest(w, "couldn't parse requestBody with metrics :"+err.Error())
		return
	}
	for _, metric := range requestBodySerialised {
		audit.AddMetric(r.Context(), metric.ID)
	}

	err = h.service.UpdateMetrics(r.Context(), requestBodySerialised)
	if err != nil {
		h.logger.Error("failed to update metrics: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
