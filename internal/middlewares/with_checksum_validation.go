package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	"go.uber.org/zap"
)

func WithChecksumValidation(key string, logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(helpers.ShaSimpleSignatureHeader) == "" {
				logger.Debug("checksum validation header is empty")
				next.ServeHTTP(w, r)
				return
			}

			payloadData, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("failed to read request body", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(payloadData))
			payloadHash := helpers.Sha256WithKeyHex(payloadData, []byte(key))
			requestShaHeaderValue := r.Header.Get(helpers.ShaSimpleSignatureHeader)

			if payloadHash != requestShaHeaderValue {
				logger.Warn("invalid signature provided")
				http.Error(w, "invalid signature", http.StatusBadRequest)
				return
			}

			logger.Info("successfully validated signature")
			next.ServeHTTP(w, r)
		})

	}
}
