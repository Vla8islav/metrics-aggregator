package main

import (
	"context"
	"log"
	"os"

	"github.com/Vla8islav/metrics-aggregator/internal/agent"
	"github.com/Vla8islav/metrics-aggregator/internal/config"
)

func main() {
	currentConfig := config.ReadFlagsClient(os.Args[1:])

	ag := agent.NewAgent(currentConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverAddr := "http://" + currentConfig.ServerAddress.Value
	pollInterval := currentConfig.PollInterval.Duration
	reportInterval := currentConfig.ReportInterval.Duration
	secret := currentConfig.SecretKey
	log.Printf("agent started: metric_poll=%s report=%s server=%s secret=%s",
		pollInterval, reportInterval, serverAddr, secret)
	ag.Start(ctx)

}
