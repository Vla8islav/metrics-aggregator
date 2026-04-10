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
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/Vla8islav/metrics-aggregator/internal/service"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	currentConfig := config.ReadFlags(os.Args[1:])
	logger.Info("starting server ", zap.String("Server addr", currentConfig.ServerAddress.Value))

	var db domain.MetricRepository
	if currentConfig.DatabaseDSN.BeenSet {
		db, err = repository.NewPostgresStorage(currentConfig, currentConfig.MigrationsFolder.Value)
		if err != nil {
			logger.Fatal("failed to initialize metrics repository", zap.Error(err))
			return
		}
	} else if !currentConfig.Restore.BeenSet || !currentConfig.Restore.Value {
		logger.Info("making a fresh in-memory DB because the restore flag is false")
		db = repository.NewMemStorage(currentConfig)
	} else if currentConfig.Restore.Value {
		logger.Info("trying to load data from file into the in-memory DB")
		dbInMemory := repository.NewMemStorage(currentConfig)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go dbInMemory.RunSaver(ctx)
		err = dbInMemory.Restore(ctx)
		if err != nil {
			logger.Fatal("failed to restore database", zap.Error(err))
		}
		db = dbInMemory
	} else {
		logger.Fatal("something strange happened: " +
			"restore and connection string parameters are incorrect. Exiting...")
		return
	}

	srvApp := service.NewMetricsService(db)
	h := handler.NewHandler(srvApp, logger)
	r := handler.NewRouter(h)

	handlerWithMW := middlewares.ChainMiddlewares(
		r,
		middlewares.WithLogging(logger),
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
