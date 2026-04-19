package middlewares

import (
	"bytes"
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
)

// the first Write call locks Headers, so we buffer response until the last moment
type capturingSignResponseWriter struct {
	http.ResponseWriter
	statusCode     int
	bodyCopy       *bytes.Buffer
	headerInstance http.Header
	signKey        string
}

func newCapturingSignResponseWriter(w http.ResponseWriter, signKey string) *capturingSignResponseWriter {
	return &capturingSignResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		bodyCopy:       bytes.NewBuffer(nil),
		headerInstance: make(http.Header),
		signKey:        signKey,
	}
}

func (w *capturingSignResponseWriter) Header() http.Header {
	return w.headerInstance
}

func (w *capturingSignResponseWriter) Write(b []byte) (int, error) {
	return w.bodyCopy.Write(b)
}

func (w *capturingSignResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

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
