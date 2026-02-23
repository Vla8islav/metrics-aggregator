package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handler.GetMetrics)

	srv := &http.Server{Addr: ":8080", Handler: r, ReadTimeout: 5 * time.Second}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return
	}

}
