package handler

import (
	"fmt"
	"log"
	"net/http"
	"sort"
)

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeMethodNotAllowed(w, "only GET method is allowed")
		return
	}

	metricsExport, err := h.service.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting all metrics: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	// Laziest HTML page ever
	_, err = w.Write([]byte("<HTML>"))
	if err != nil {
		log.Printf("error opening the HTML tag: %s", err.Error())
	}
	var htmlStrings []string
	for k, v := range metricsExport.Gauges {
		htmlStrings = append(htmlStrings, fmt.Sprintf("%s %f </br>", k, v))
	}

	for k, v := range metricsExport.Counters {
		htmlStrings = append(htmlStrings, fmt.Sprintf("%s %d </br>", k, v))
	}
	sort.Strings(htmlStrings)

	for _, str := range htmlStrings {
		_, err = w.Write([]byte(str))
		if err != nil {
			return
		}
	}

	_, err = w.Write([]byte("</HTML>"))
	if err != nil {
		log.Printf("error closing the HTML tag: %s", err.Error())
	}

}
