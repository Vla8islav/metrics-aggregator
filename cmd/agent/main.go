package main

import (
	"context"
	"log"
	"os"

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

	currentConfig := config.ReadFlagsClient(os.Args[1:])

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ag := agent.NewAgent(currentConfig, logger)

	serverAddr := "http://" + currentConfig.ServerAddress.Value
	pollInterval := currentConfig.PollInterval.Duration
	reportInterval := currentConfig.ReportInterval.Duration
	log.Printf("agent started: metric_poll=%s report=%s server=%s secret_is_set=%v",
		pollInterval, reportInterval, serverAddr, currentConfig.SecretKey.BeenSet)
	ag.Start(ctx)

}
