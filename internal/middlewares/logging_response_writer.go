package middlewares

import "net/http"

// responseData stores response details captured by loggingResponseWriter.
type (
	responseData struct {
		status int
		size   int
		body   []byte
	}

	// loggingResponseWriter wraps http.ResponseWriter and records response metadata.
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

// Write sends b to the underlying response writer and records the number of bytes written.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

// WriteHeader sends statusCode to the underlying response writer and records it.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
