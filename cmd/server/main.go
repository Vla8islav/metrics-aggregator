package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/audit"
	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/Vla8islav/metrics-aggregator/internal/middlewares"
	"github.com/Vla8islav/metrics-aggregator/internal/service"
	"go.uber.org/zap"
)

import _ "net/http/pprof"

func main() {

	//< for testing only, delete in prod
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

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

	var sinks []audit.Sink
	if currentConfig.AuditFile.BeenSet {
		fileSink := audit.NewFileSink(currentConfig.AuditFile.Value)
		sinks = append(sinks, fileSink)
	}
	if currentConfig.AuditURL.BeenSet {
		webSink, err := audit.NewWebSink(currentConfig.AuditURL.Value)
		if err != nil {
			logger.Fatal("failed to initialize web sink", zap.Error(err))
		}
		sinks = append(sinks, webSink)
	}
	publisher := audit.NewPublisher(sinks...)

	handlerWithMW := middlewares.ChainMiddlewares(
		r,
		middlewares.WithLogging(logger),
		middlewares.WithAudit(publisher),
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

	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()
	err = srvImpl.ListenAndServe()

	if err != nil {
		logger.Fatal(err.Error())
		return
	}
}
