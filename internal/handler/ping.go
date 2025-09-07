package handler

import "net/http"

func (m metricHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := m.service.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
