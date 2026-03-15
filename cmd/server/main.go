package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/Vla8islav/metrics-aggregator/internal/service"
)

func main() {

	db := repository.NewMemStorage()
	srvApp := service.NewMetricsService(db)
	h := handler.NewHandler(srvApp)
	r := handler.NewRouter(h)

	srvImpl := &http.Server{Addr: config.ReadFlags().ServerAddress,
		Handler:     r,
		ReadTimeout: 5 * time.Second}

	err := srvImpl.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return
	}

}
