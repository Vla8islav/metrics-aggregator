package handler

import (
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) DbPing(w http.ResponseWriter, r *http.Request) {

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
		http.Error(w, "database is unavailable", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
