package helpers

import (
	"crypto/sha256"
	"encoding/hex"
)

const ShaSimpleSignatureHeader = "HashSHA256"

func Sha256WithKey(body, key []byte) []byte {
	h := sha256.New()
	h.Write(body)
	h.Write(key)
	return h.Sum(nil)
}

func Sha256WithKeyHex(body, key []byte) string {
	return hex.EncodeToString(Sha256WithKey(body, key))
}
