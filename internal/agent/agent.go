package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	"github.com/Vla8islav/metrics-aggregator/internal/model"
)

type Agent struct {
	client         *helpers.HTTPRetryClient
	serverAddr     string
	pollInterval   time.Duration
	reportInterval time.Duration

	gauges *models.Stats
	config *config.OptionsClient
}

func NewAgent(currentConfig *config.OptionsClient) *Agent {
	serverAddr := "http://" + currentConfig.ServerAddress.Value
	pollInterval := currentConfig.PollInterval.Duration
	reportInterval := currentConfig.ReportInterval.Duration

	s := models.NewStats()
	retryClient := helpers.NewHTTPRetryClient(helpers.DefaultShouldRetryStatus,
		5*time.Second, 2)
	return &Agent{
		client:         retryClient,
		serverAddr:     serverAddr,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		gauges:         s,
		config:         currentConfig,
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

func (a *Agent) report(ctx context.Context) error {
	payload := make([]models.Metrics, 0)
	// send gauges
	for name, value := range a.gauges.GetGauges() {
		payload = append(payload, models.Metrics{MType: models.Gauge, ID: name, Value: &value})
	}

	// send counters
	for name, value := range a.gauges.GetCounters() {
		payload = append(payload, models.Metrics{MType: models.Counter, ID: name, Delta: &value})
	}

	err := a.sendBatch(ctx, payload)
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) send(ctx context.Context, metricType models.MetricType, metricName string, gauge *float64, counter *int64) error {
	if gauge == nil && counter == nil {
		return fmt.Errorf("both gauge and counter are nil")
	}
	// POST http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	base, err := url.Parse(a.serverAddr)
	if err != nil {
		return err
	}

	base.Path = path.Join(
		base.Path,
		"update",
	)

	payload := models.Metrics{
		ID:    metricName,
		MType: metricType,
		Delta: counter,
		Value: gauge,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	payloadBytesCompressed, err := helpers.GzipCompress(payloadBytes)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base.String(), bytes.NewReader(payloadBytesCompressed))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if gauge != nil {
			return fmt.Errorf("server returned %s for %s %s=%f", resp.Status, metricType, metricName, *gauge)
		}

		return fmt.Errorf("server returned %s for %s %s=%v", resp.Status, metricType, metricName, *counter)
	}
	return nil
}

func (a *Agent) getSignatureHeaderValue(payloadBytes []byte) string {
	return helpers.Sha256WithKeyHex(payloadBytes, []byte(a.config.SecretKey.Value))
}

func (a *Agent) sendBatch(ctx context.Context, metrics []models.Metrics) error {
	if metrics == nil {
		return nil
	}
	// POST http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	base, err := url.Parse(a.serverAddr)
	if err != nil {
		return fmt.Errorf("couldn't parse server addr: %w", err)
	}

	base.Path = path.Join(
		base.Path,
		"updates",
	)

	payloadBytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("couldn't marshal metrics payload: %w", err)
	}

	payloadBytesCompressed, err := helpers.GzipCompress(payloadBytes)
	if err != nil {
		return fmt.Errorf("couldn't compress metrics payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		base.String(), bytes.NewReader(payloadBytesCompressed))

	if err != nil {
		return fmt.Errorf("couldn't build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	signatureHeaderValue := a.getSignatureHeaderValue(payloadBytes)
	req.Header.Set(helpers.ShaSimpleSignatureHeader, signatureHeaderValue)

	resp, err := a.client.Do(req)

	if err != nil {
		return fmt.Errorf("couldn't make a request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s for %s payload", resp.Status, string(payloadBytes))
	}
	return nil
}
