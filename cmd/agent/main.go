package main

import (
	"context"
	"log"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/agent"
)

func main() {
	serverAddr := "http://localhost:8080"
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	ag := agent.NewAgent(serverAddr, pollInterval, reportInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("agent started: metric_poll=%s report=%s server=%s", pollInterval, reportInterval, serverAddr)
	ag.Start(ctx)

}
