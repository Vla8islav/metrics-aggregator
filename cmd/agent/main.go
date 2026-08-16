package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vla8islav/metrics-aggregator/internal/agent"
	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"go.uber.org/zap"
)

func main() {
	printBuildInfo()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	currentConfig := config.ReadFlagsClient(os.Args[1:], logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	ag, err := agent.NewAgent(currentConfig, logger)
	if err != nil {
		log.Fatalf("failed to initialize agent: %v", err)
	}

	serverAddr := "http://" + currentConfig.ServerAddress.Value
	pollInterval := currentConfig.PollInterval.Duration
	reportInterval := currentConfig.ReportInterval.Duration

	logger.Info("agent started",
		zap.Duration("metric_poll", pollInterval),
		zap.Duration("report", reportInterval),
		zap.String("server", serverAddr),
		zap.Bool("crypto_key_is_set", currentConfig.CryptoKey.BeenSet),
		zap.Bool("secret_is_set", currentConfig.SecretKey.BeenSet),
	)

	ag.Start(ctx)
	logger.Info("shutdown signal recieved")
	stop()
	logger.Info("agent stopped")
}
