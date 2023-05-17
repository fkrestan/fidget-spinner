package middleware

import (
	"net/http"
)

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func newStatusWriter(w http.ResponseWriter) *statusWriter {
	return &statusWriter{w, http.StatusOK}
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
