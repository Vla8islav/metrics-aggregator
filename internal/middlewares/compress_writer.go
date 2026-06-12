package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// gzipWriter wraps a ResponseWriter and writes response bodies through a gzip writer
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write compresses b before writing it to the underlying response
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// WithGzipCompression returns mw that decompresses gzip request bodies and compresses gzip-responses
func WithGzipCompression() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !handleInboundCompression(w, r) {
				return
			}

			handleOutgoingCompression(w, r, next)
		})

	}
}

// gzipPool reuses gzip writers between requests to reduce allocations
// was part of the memory optimisation
var gzipPool = sync.Pool{
	New: func() any {
		// NewWriterLevel only errors on an invalid level —
		// impossible with the gzip.BestSpeed constant, so safe to swallow.
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return w
	},
}

// handleInboundCompression replaces a gzip-compressed request body with a decompressed reader
func handleInboundCompression(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return false
		}
		r.Body = gzipReader
		r.Header.Del("Content-Encoding")
		r.Header.Del("Content-Length")
	}
	return true
}

// handleOutgoingCompression compresses the response when the client accepts gzip encoding
func handleOutgoingCompression(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// проверяем, что клиент поддерживает gzip-сжатие
	// это упрощённый пример. В реальном приложении следует проверять все
	// значения r.Header.Values("Accept-Encoding") и разбирать строку
	// на составные части, чтобы избежать неожиданных результатов
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		// если gzip не поддерживается, передаём управление
		// дальше без изменений
		next.ServeHTTP(w, r)
		return
	}

	// создаём gzip.Writer поверх текущего w
	gz := gzipPool.Get().(*gzip.Writer)
	gz.Reset(w) // reuse
	defer func() {
		gz.Close()
		gzipPool.Put(gz)
	}()

	w.Header().Set("Content-Encoding", "gzip")
	// передаём обработчику страницы переменную типа gzipWriter для вывода данных
	next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)

}
