package helpers

import (
	"crypto/sha256"
	"encoding/hex"
)

// ShaSimpleSignatureHeader is the HTTP header name used for SHA-256 request signatures
const ShaSimpleSignatureHeader = "HashSHA256"

// Sha256WithKey returns the SHA-256 digest of body followed by key
func Sha256WithKey(body, key []byte) []byte {
	h := sha256.New()
	h.Write(body)
	h.Write(key)
	return h.Sum(nil)
}

// Sha256WithKeyHex returns the hex-encoded SHA-256 digest of body followed by key
func Sha256WithKeyHex(body, key []byte) string {
	return hex.EncodeToString(Sha256WithKey(body, key))
}
