package main

import (
	"context"
	"log"

	"github.com/Vla8islav/metrics-aggregator/internal/agent"
	"github.com/Vla8islav/metrics-aggregator/internal/config"
)

func main() {
	currentConfig := config.ReadFlags()
	serverAddr := "http://" + currentConfig.ServerAddress.Value
	pollInterval := currentConfig.PollInterval.Duration
	reportInterval := currentConfig.ReportInterval.Duration

	ag := agent.NewAgent(serverAddr, pollInterval, reportInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("agent started: metric_poll=%s report=%s server=%s", pollInterval, reportInterval, serverAddr)
	ag.Start(ctx)

}
