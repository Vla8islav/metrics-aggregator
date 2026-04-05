package handler

import (
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) writeBadRequest(w http.ResponseWriter, msg string) {
	h.logger.Error("bad request", zap.String("msg", msg))
	http.Error(w, msg, http.StatusBadRequest)
}

func (h *Handler) writeMethodNotAllowed(w http.ResponseWriter, msg string) {
	h.logger.Error("method not allowed: ", zap.String("msg", msg))
	http.Error(w, msg, http.StatusMethodNotAllowed)
}
