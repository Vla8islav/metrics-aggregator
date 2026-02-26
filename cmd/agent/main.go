package main

import (
	"context"
	"log"

	"github.com/Vla8islav/metrics-aggregator/internal/agent"
	"github.com/Vla8islav/metrics-aggregator/internal/config"
)

func main() {
	serverAddr := "http://" + config.ReadFlags().ServerAddress
	pollInterval := config.ReadFlags().PollInterval
	reportInterval := config.ReadFlags().ReportInterval

	ag := agent.NewAgent(serverAddr, pollInterval, reportInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("agent started: metric_poll=%s report=%s server=%s", pollInterval, reportInterval, serverAddr)
	ag.Start(ctx)

}
