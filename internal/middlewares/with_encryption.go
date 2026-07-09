package middlewares

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/model"
	"go.uber.org/zap"
)

// WithEncryption returns middleware that tries to decrypt the message when the key isn't empty
func WithEncryption(privateKey *rsa.PrivateKey, logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil || r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}
			decryptedBody, err := handleInboundEncryption(r, privateKey)
			if err != nil {
				logger.Warn("failed to decrypt request body", zap.Error(err))
				http.Error(w, "failed to decrypt request body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(decryptedBody))
			r.ContentLength = int64(len(decryptedBody))
			next.ServeHTTP(w, r)
		})
	}
}

func handleInboundEncryption(r *http.Request, privateKey *rsa.PrivateKey) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var payload models.EncryptedPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, payload.Key)
	if err != nil {
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

	decryptedBody, err := aesGCM.Open(nil, payload.Nonce, payload.Data, nil)
	if err != nil {
		return nil, err
	}

	return decryptedBody, nil
}
