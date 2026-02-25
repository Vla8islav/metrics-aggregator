package handler

import (
	"fmt"
	"log"
	"net/http"
)

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	metricsExport, err := h.repo.GetAll()
	if err != nil {
		log.Printf("Error getting all metrics: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("<HTML>"))
	for k, v := range metricsExport.Gauges {
		_, err = w.Write([]byte(fmt.Sprintf("%s %f </br>", k, v)))
		if err != nil {
			return
		}
	}

	for k, v := range metricsExport.Counters {
		_, err = w.Write([]byte(fmt.Sprintf("%s %d </br>", k, v)))
		if err != nil {
			return
		}
	}
	_, err = w.Write([]byte("</HTML>"))

}
