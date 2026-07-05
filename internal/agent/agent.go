// Package agent periodically collects runtime metrics and reports them to the server
package agent

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"go.uber.org/zap"
)

// Agent periodically collects runtime metrics and reports them to the server
// generate:reset
type Agent struct {
	client         *helpers.HTTPRetryClient
	serverAddr     string
	pollInterval   time.Duration
	reportInterval time.Duration
	rateLimit      int

	gauges *models.Stats
	config *config.OptionsClient
	logger *zap.Logger
}

// encryptedPayload contains a hybrid-encrypted request body.
// it's necessary to send large payloads
// Data contains the request body encrypted with AES-GCM
type encryptedPayload struct {
	Key   []byte `json:"key"`
	Nonce []byte `json:"nonce"`
	Data  []byte `json:"data"`
}

// NewAgent creates an Agent configured with the provided client options and logger
func NewAgent(currentConfig *config.OptionsClient, logger *zap.Logger) *Agent {
	serverAddr := "http://" + currentConfig.ServerAddress.Value
	pollInterval := currentConfig.PollInterval.Duration
	reportInterval := currentConfig.ReportInterval.Duration
	rateLimit := currentConfig.RateLimit.Value
	if rateLimit <= 0 {
		rateLimit = 1
	}

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
		logger:         logger,
		rateLimit:      rateLimit,
	}
}

// Start begins metric collection and reporting until ctx is canceled
func (a *Agent) Start(ctx context.Context) {

	err := a.gauges.Update()
	if err != nil {
		a.logger.Error("failed to update gauges", zap.Error(err))
		return
	}

	jobs := make(chan job, a.rateLimit)
	results := make(chan result)

	var workersWG sync.WaitGroup
	for i := 1; i <= a.rateLimit; i++ {
		workersWG.Add(1)
		go a.workerReport(ctx, i, jobs, results, &workersWG)
	}

	var resultsWG sync.WaitGroup
	resultsWG.Add(1)

	go func() {
		defer resultsWG.Done()
		a.runMetricsGatherer(ctx)
	}()

	resultsDone := make(chan struct{})

	go func() {
		for res := range results {
			if res.Err != nil {
				a.logger.Warn("report job failed", zap.Int("jobID", res.JobID), zap.Error(res.Err))
			}
		}
		close(resultsDone)
	}()

	a.runReporter(ctx, jobs)

	close(jobs)

	workersWG.Wait()
	close(results)
	<-resultsDone
	resultsWG.Wait()
}

// runMetricsGatherer periodically refreshes the agent's in-memory metric values
func (a *Agent) runMetricsGatherer(ctx context.Context) {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.gauges.Update(); err != nil {
				a.logger.Warn("failed to gather metrics", zap.Error(err))
				continue
			}

		}
	}
}

// job represents a scheduled reporting task
type job struct {
	ID int
}

// result contains the outcome of a reporting job
type result struct {
	JobID int
	Value string
	Err   error
}

// workerReport processes reporting jobs and publishes their results
func (a *Agent) workerReport(ctx context.Context, id int, jobs <-chan job, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		err := a.report(ctx)
		if err == nil {
			a.logger.Debug("report job finished", zap.Int("worker", id), zap.Int("jobID", j.ID))
		}
		select {
		case <-ctx.Done():
			return

		case results <- result{
			JobID: j.ID,
			Value: fmt.Sprintf("workerReport %d processed j %d", id, j.ID),
			Err:   err,
		}: // empty
		}
	}
}

// report builds a metrics payload from the current stats and sends it to the server
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

// send sends a single metric update to the server
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

// getSignatureHeaderValue returns the request signature for payloadBytes
func (a *Agent) getSignatureHeaderValue(payloadBytes []byte) string {
	return helpers.Sha256WithKeyHex(payloadBytes, []byte(a.config.SecretKey.Value))
}

func (a *Agent) readPublicKey() (*x509.Certificate, error) {
	if !a.config.CryptoKey.BeenSet || "" == a.config.CryptoKey.Value {
		return nil, fmt.Errorf("crypto key wasn't set")
	}

	publicKeyPath := a.config.CryptoKey.Value

	certificateBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	certificatePemBlock, _ := pem.Decode(certificateBytes)
	if certificatePemBlock == nil {
		return nil, errors.New("certificate not found")
	}

	certificate, err := x509.ParseCertificate(certificatePemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

// encryptMessage encrypts the payload using hybrid encryption.
//
// RSA can encrypt only small messages, so the payload itself is encrypted with AES-GCM
// The random AES key is then encrypted with the server RSA public key from the certificate
func (a *Agent) encryptMessage(message []byte) ([]byte, error) {
	certificate, err := a.readPublicKey()
	if err != nil {
		return nil, err
	}

	publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("certificate public key is not RSA")
	}

	aesKey := make([]byte, 32)
	if _, err = rand.Read(aesKey); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	encryptedData := aesGCM.Seal(nil, nonce, message, nil)
	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, aesKey)
	if err != nil {
		return nil, err
	}

	payload := encryptedPayload{
		Key:   encryptedKey,
		Nonce: nonce,
		Data:  encryptedData,
	}

	return json.Marshal(payload)
}

// sendBatch sends a gzip-compressed and signed metrics batch to the server
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
	a.logger.Info("sending batch payload", zap.String("payload", string(payloadBytes)))

	// payload encryption using the cert
	if a.config.CryptoKey.BeenSet {
		payloadBytes, err = a.encryptMessage(payloadBytes)
		if err != nil {
			return fmt.Errorf("couldn't encrypt the payload %w", err)
		}
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
		return fmt.Errorf("server returned %s for %s payload", resp.Status, string(payloadBytes)[:40])
	}
	return nil
}

// runReporter schedules reporting jobs at the configured report interval
func (a *Agent) runReporter(ctx context.Context, jobs chan<- job) {
	ticker := time.NewTicker(a.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case <-ctx.Done():
				return
			case jobs <- job{ID: int(time.Now().UnixNano())}:
			}
		}
	}
}
