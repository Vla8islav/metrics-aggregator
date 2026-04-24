package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/Vla8islav/metrics-aggregator/internal/middlewares"
	"github.com/Vla8islav/metrics-aggregator/internal/service"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	currentConfig := config.ReadFlagsServer(os.Args[1:])
	logger.Info("starting server ", zap.String("Server addr", currentConfig.ServerAddress.Value))

	var db domain.MetricRepository
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err = initDB(ctx, currentConfig, logger)
	if err != nil {
		logger.Fatal("failed to initialize db", zap.Error(err))
	}

	srvApp := service.NewMetricsService(db)
	h := handler.NewHandler(srvApp, logger)
	r := handler.NewRouter(h)

	handlerWithMW := middlewares.ChainMiddlewares(
		r,
		middlewares.WithLogging(logger),
	)

	if currentConfig.SecretKey.BeenSet {
		handlerWithMW = middlewares.ChainMiddlewares(
			handlerWithMW,
			middlewares.WithChecksum(currentConfig.SecretKey.Value, logger),
		)
	}
	// compression should come last
	handlerWithMW = middlewares.ChainMiddlewares(
		handlerWithMW,
		middlewares.WithGzipCompression(),
	)

	srvImpl := &http.Server{Addr: currentConfig.ServerAddress.Value,
		Handler:      handlerWithMW,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	err = srvImpl.ListenAndServe()
	if err != nil {
		logger.Fatal(err.Error())
		return
	}
}
