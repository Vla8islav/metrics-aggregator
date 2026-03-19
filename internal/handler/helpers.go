package handler

import (
	"log"
	"net/http"
)

func writeBadRequest(w http.ResponseWriter, msg string) {
	log.Println(msg)
	http.Error(w, msg, http.StatusBadRequest)
}
