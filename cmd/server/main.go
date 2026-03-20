package main

import (
	"net/http"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/Vla8islav/metrics-aggregator/internal/middlewares"
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/Vla8islav/metrics-aggregator/internal/service"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // flushes buffer, if any

	db := repository.NewMemStorage()
	srvApp := service.NewMetricsService(db)
	h := handler.NewHandler(srvApp)
	r := handler.NewRouter(h)

	handlerWithMW := middlewares.ChainMiddlewares(
		r,
		middlewares.WithLogging(logger),
		middlewares.WithGzipCompression(),
	)

	srvImpl := &http.Server{Addr: config.ReadFlags().ServerAddress,
		Handler:     handlerWithMW,
		ReadTimeout: 5 * time.Second}

	err = srvImpl.ListenAndServe()
	if err != nil {
		logger.Fatal(err.Error())
		return
	}

}
