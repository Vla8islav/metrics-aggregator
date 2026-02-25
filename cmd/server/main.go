package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/Vla8islav/metrics-aggregator/internal/router"
)

func main() {

	db := repository.NewMemStorage()
	h := handler.NewHandler(db)
	r := router.NewRouter(h)

	srv := &http.Server{Addr: ":8080", Handler: r, ReadTimeout: 5 * time.Second}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return
	}

}
