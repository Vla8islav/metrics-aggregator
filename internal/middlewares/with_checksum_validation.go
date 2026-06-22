package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	"go.uber.org/zap"
)

// WithChecksum returns middleware that validates signed requests and signs responses.
//
// Requests with the checksum header are rejected when their signature doesn't match req body
// Responses are buffered, signed, and then written to the client
func WithChecksum(key string, logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !handleInboundValidation(w, r, key, logger) {
				logger.Info("inbound validation failure")
				return
			}

			err := handleOutgoingSigning(w, r, key, next)
			if err != nil {
				logger.Warn("outgoing validation failure", zap.Error(err))
			}

		})

	}
}

// handleInboundValidation validates the request checksum when the checksum header is present.
func handleInboundValidation(w http.ResponseWriter, r *http.Request, key string, logger *zap.Logger) bool {
	if r.Header.Get(helpers.ShaSimpleSignatureHeader) == "" {
		logger.Debug("checksum validation header is empty")
		return true
	}

	payloadData, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	r.Body = io.NopCloser(bytes.NewReader(payloadData))

	payloadHash := helpers.Sha256WithKeyHex(payloadData, []byte(key))
	requestShaHeaderValue := r.Header.Get(helpers.ShaSimpleSignatureHeader)

	if payloadHash != requestShaHeaderValue {
		logger.Warn("invalid signature provided, calculated hash: " + payloadHash +
			" doesn't match provided hash: " + requestShaHeaderValue +
			" secret " + key)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return false
	}

	logger.Info("successfully validated signature")
	return true
}

// handleOutgoingSigning captures the response body, signs it, and writes it to the client.
func handleOutgoingSigning(w http.ResponseWriter, r *http.Request, key string, next http.Handler) error {

	capturingWriter := newCapturingSignResponseWriter(w, key)

	next.ServeHTTP(capturingWriter, r)

	err := capturingWriter.FlushToOriginal()
	if err != nil {
		return err
	}
	return nil
}
