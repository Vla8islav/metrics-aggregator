package handler

import (
	"net/http"

	"go.uber.org/zap"
)

// DBPing handles database health checks
//
// accepts only GET requests, returns 200 when the storage layer is available
func (h *Handler) DBPing(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		h.logger.Warn("method not allowed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.service.Ping(r.Context())
	if err != nil {
		h.logger.Error("db ping failed",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
