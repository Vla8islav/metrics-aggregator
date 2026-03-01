package agent

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/handler"
	"github.com/Vla8islav/metrics-aggregator/internal/model"
)

type Agent struct {
	client         *http.Client
	serverAddr     string
	pollInterval   time.Duration
	reportInterval time.Duration

	gauges *models.Stats
}

func NewAgent(serverAddr string, pollInterval, reportInterval time.Duration) *Agent {
	s := models.NewStats()
	return &Agent{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		serverAddr:     serverAddr,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		gauges:         s,
	}
}

func (a *Agent) Start(ctx context.Context) {

	// init the tickers
	pollTicker := time.NewTicker(a.pollInterval)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(a.reportInterval)
	defer reportTicker.Stop()

	err := a.gauges.Update()
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			err = a.gauges.Update()
			if err != nil {
				return
			}
		case <-reportTicker.C:
			a.report(ctx)
		}
	}

}

func formatFloat(v float64) string {
	// grooming floats a bit
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func (a *Agent) report(ctx context.Context) error {
	// send gauges
	for name, value := range a.gauges.GetGauges() {
		if err := a.send(ctx, handler.Gauge, name, formatFloat(value)); err != nil {
			return err
		}
	}

	// send counters
	for name, value := range a.gauges.GetCounters() {
		if err := a.send(ctx, handler.Counter, name, strconv.FormatInt(value, 10)); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) send(ctx context.Context, metricType handler.MetricType, metricName, metricValue string) error {
	// POST http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	url := path.Join(a.serverAddr, "update", string(metricType), metricName, metricValue)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s for %s %s=%s", resp.Status, metricType, metricName, metricValue)
	}
	return nil
}
