package models

// EncryptedPayload contains a hybrid-encrypted request body.
// it's necessary to send large payloads
// Data contains the request body encrypted with AES-GCM
type EncryptedPayload struct {
	Key   []byte `json:"key"`
	Nonce []byte `json:"nonce"`
	Data  []byte `json:"data"`
}
