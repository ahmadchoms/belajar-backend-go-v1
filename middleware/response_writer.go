package middleware

import "net/http"

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
