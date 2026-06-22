package middlewares

import (
	"bytes"
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
)

// capturingSignResponseWriter buffers a response so it can be signed before headers are written.
//
// The first write to the original ResponseWriter locks its headers, so this writer stores
// headers and body data until FlushToOriginal copies them to the wrapped writer
type capturingSignResponseWriter struct {
	http.ResponseWriter
	statusCode     int
	bodyCopy       *bytes.Buffer
	headerInstance http.Header
	signKey        string
}

// newCapturingSignResponseWriter creates a response writer that signs the buffered response body.
func newCapturingSignResponseWriter(w http.ResponseWriter, signKey string) *capturingSignResponseWriter {
	return &capturingSignResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		bodyCopy:       bytes.NewBuffer(nil),
		headerInstance: make(http.Header),
		signKey:        signKey,
	}
}

// Header returns the buffered response headers.
func (w *capturingSignResponseWriter) Header() http.Header {
	return w.headerInstance
}

// Write appends b to the buffered response body.
func (w *capturingSignResponseWriter) Write(b []byte) (int, error) {
	return w.bodyCopy.Write(b)
}

// WriteHeader stores statusCode until the response is flushed.
func (w *capturingSignResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// FlushToOriginal signs the buffered body and writes the full response to the wrapped writer.
func (w *capturingSignResponseWriter) FlushToOriginal() error {
	signature := helpers.Sha256WithKeyHex(w.bodyCopy.Bytes(), []byte(w.signKey))
	w.headerInstance.Set(helpers.ShaSimpleSignatureHeader, signature)

	originalHeader := w.ResponseWriter.Header()
	for key, values := range w.headerInstance {
		for _, value := range values {
			originalHeader.Add(key, value)
		}
	}

	w.ResponseWriter.WriteHeader(w.statusCode)

	_, err := w.ResponseWriter.Write(w.bodyCopy.Bytes())
	return err
}
