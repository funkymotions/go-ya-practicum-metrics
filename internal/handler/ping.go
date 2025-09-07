package handler

import (
	"net/http"
)

func (h metricHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
