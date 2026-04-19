package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	"go.uber.org/zap"
)

func WithChecksum(key string, logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !handleInboundValidation(w, r, key, logger) {
				logger.Info("inbound validation failure")
				return
			}

			handleOutgoingSigning(w, r, key, next)

		})

	}
}

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
			" payload " + string(payloadData) +
			" secret " + key)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return false
	}

	logger.Info("successfully validated signature")
	return true
}

func handleOutgoingSigning(w http.ResponseWriter, r *http.Request, key string, next http.Handler) {

	capturingWriter := newCapturingSignResponseWriter(w, key)

	w.Header().Set("Custom-Header-Before", "blahblahblah before original")
	next.ServeHTTP(capturingWriter, r) // Сначала вызываем следующий обработчик, чтобы записать данные в буфер

	// Теперь вычисляем подпись и добавляем её в заголовок
	signature := helpers.Sha256WithKeyHex(capturingWriter.bodyCopy.Bytes(), []byte(key))
	w.Header().Set(helpers.ShaSimpleSignatureHeader, signature)
	w.Header().Set("Custom-Header-After", "blahblahblah after original")

}
