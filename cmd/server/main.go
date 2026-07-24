package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/audit"
	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	"github.com/Vla8islav/metrics-aggregator/internal/grpcserver"
	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	"github.com/Vla8islav/metrics-aggregator/internal/middlewares"
	"github.com/Vla8islav/metrics-aggregator/internal/middlewares_grpc"
	"github.com/Vla8islav/metrics-aggregator/internal/proto"
	"github.com/Vla8islav/metrics-aggregator/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	_ "net/http/pprof"
)

func main() {

	printBuildInfo()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	currentConfig, err := config.ReadFlagsServer(os.Args[1:])
	if err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
	}
	logger.Info("starting server ", zap.String("Server addr", currentConfig.ServerAddress.Value))

	var db domain.MetricRepository
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	db, err = initDB(ctx, currentConfig, logger)
	if err != nil {
		logger.Fatal("failed to initialize db", zap.Error(err))
	}

	srvApp := service.NewMetricsService(db)

	// <grpc>
	ipChecker, err := middlewares_grpc.WithIPChecker(
		currentConfig.TrustedSubnets,
		logger,
	)
	if err != nil {
		logger.Fatal(
			"failed to initialize gRPC IP checker",
			zap.Error(err),
		)
	}

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(middlewares_grpc.WithLogging(logger),
			ipChecker,
		))

	proto.RegisterMetricsServer(
		grpcSrv,
		grpcserver.NewGRPCServer(srvApp),
	)
	grpcListener, err := net.Listen(
		"tcp",
		currentConfig.ServerAddressGRPC.Value,
	)

	if err != nil {
		logger.Fatal("failed to create grpc listener", zap.Error(err))
	}

	go func() {
		logger.Info("starting grpc server ", zap.String("Server addr", currentConfig.ServerAddressGRPC.Value))
		if err = grpcSrv.Serve(grpcListener); err != nil {
			logger.Error("grpc server stopped", zap.Error(err))
		}
	}()
	// </grpc>

	h := handler.NewHandler(srvApp, logger)
	r := handler.NewRouter(h)

	var sinks []audit.Sink
	if currentConfig.AuditFile.BeenSet {
		fileSink := audit.NewFileSink(currentConfig.AuditFile.Value)
		defer fileSink.Close()
		sinks = append(sinks, fileSink)
	}
	if currentConfig.AuditURL.BeenSet {
		webSink, err2 := audit.NewWebSink(currentConfig.AuditURL.Value)
		if err2 != nil {
			logger.Fatal("failed to initialize web sink", zap.Error(err))
		}
		sinks = append(sinks, webSink)
	}
	publisher := audit.NewPublisher(sinks...)

	handlerWithMW := middlewares.ChainMiddlewares(
		r,
		middlewares.WithLogging(logger),
		middlewares.WithAudit(publisher),
		middlewares.WithIpChecker(currentConfig.TrustedSubnets, logger),
	)

	if currentConfig.SecretKey.BeenSet {
		handlerWithMW = middlewares.ChainMiddlewares(
			handlerWithMW,
			middlewares.WithChecksum(currentConfig.SecretKey.Value, logger),
		)
	}

	if currentConfig.CryptoKey.BeenSet {
		privateKey, err2 := helpers.ReadPrivateKey(currentConfig.CryptoKey.Value)
		if err2 != nil {
			logger.Fatal("failed to read private key", zap.Error(err))
		}
		handlerWithMW = middlewares.ChainMiddlewares(
			handlerWithMW,
			middlewares.WithEncryption(privateKey, logger),
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

	errCh := make(chan error, 1)
	go func() {
		errCh <- srvImpl.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err = <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed", zap.Error(err))
		}
	}

	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	grpcSrv.GracefulStop()
	if err = srvImpl.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("failed to shutdown server gracefully", zap.Error(err))
	}

	if err = <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("server failed during shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}
