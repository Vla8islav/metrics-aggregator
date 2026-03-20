package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Честно спер из примеров
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

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
	gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	defer gz.Close()

	w.Header().Set("Content-Encoding", "gzip")
	// передаём обработчику страницы переменную типа gzipWriter для вывода данных
	next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
}
