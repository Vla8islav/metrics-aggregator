package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
)

func main() {

	db := repository.NewMemStorage()
	h := handler.NewHandler(db)
	r := handler.NewRouter(h)

	srv := &http.Server{Addr: config.ReadFlags().ServerAddress, Handler: r, ReadTimeout: 5 * time.Second}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return
	}

}
