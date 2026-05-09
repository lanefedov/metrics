package server

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	n, err := w.ResponseWriter.Write(p)
	w.size += n

	return n, err
}

func loggingMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := &loggingResponseWriter{ResponseWriter: w}

			next.ServeHTTP(lrw, r)

			if lrw.statusCode == 0 {
				lrw.statusCode = http.StatusOK
			}

			logger.Info().
				Str("uri", requestURI(r)).
				Str("method", r.Method).
				Dur("duration", time.Since(start)).
				Msg("request handled")

			logger.Info().
				Int("status", lrw.statusCode).
				Int("size", lrw.size).
				Msg("response sent")
		})
	}
}

func requestURI(r *http.Request) string {
	if r.RequestURI != "" {
		return r.RequestURI
	}

	return r.URL.RequestURI()
}
