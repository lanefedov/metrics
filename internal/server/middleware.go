package server

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
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

func (w *loggingResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
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

type gzipReadCloser struct {
	io.ReadCloser
	body io.Closer
}

func (r *gzipReadCloser) Close() error {
	readErr := r.ReadCloser.Close()
	bodyErr := r.body.Close()
	if readErr != nil {
		return readErr
	}

	return bodyErr
}

func requestDecompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !headerContainsToken(r.Header, "Content-Encoding", "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cloned := r.Clone(r.Context())
		cloned.Body = &gzipReadCloser{
			ReadCloser: gzipReader,
			body:       r.Body,
		}
		cloned.Header = r.Header.Clone()
		cloned.Header.Del("Content-Encoding")
		cloned.Header.Del("Content-Length")
		cloned.ContentLength = -1

		next.ServeHTTP(w, cloned)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	request            *http.Request
	statusCode         int
	gzipWriter         *gzip.Writer
	headerWritten      bool
	shouldCompress     bool
	compressionDecided bool
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}

	w.statusCode = statusCode
	w.decideCompression()
	w.headerWritten = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}

	if w.shouldCompress {
		return w.gzipWriter.Write(p)
	}

	return w.ResponseWriter.Write(p)
}

func (w *gzipResponseWriter) Flush() {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}

	if w.shouldCompress {
		_ = w.gzipWriter.Flush()
	}

	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *gzipResponseWriter) Close() error {
	if w.gzipWriter == nil {
		return nil
	}

	return w.gzipWriter.Close()
}

func (w *gzipResponseWriter) decideCompression() {
	if w.compressionDecided {
		return
	}
	w.compressionDecided = true

	if !headerContainsToken(w.request.Header, "Accept-Encoding", "gzip") {
		return
	}

	contentType := w.Header().Get("Content-Type")
	if !isCompressibleContentType(contentType) {
		return
	}

	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Add("Vary", "Accept-Encoding")
	w.Header().Del("Content-Length")
	w.gzipWriter = gzip.NewWriter(w.ResponseWriter)
	w.shouldCompress = true
}

func responseCompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			request:        r,
		}
		defer func() {
			_ = gzw.Close()
		}()

		next.ServeHTTP(gzw, r)
	})
}

func headerContainsToken(headers http.Header, key, token string) bool {
	for _, value := range headers.Values(key) {
		for _, part := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), token) {
				return true
			}
		}
	}

	return false
}

func isCompressibleContentType(contentType string) bool {
	base, _, _ := strings.Cut(contentType, ";")
	switch strings.TrimSpace(strings.ToLower(base)) {
	case "application/json", "text/html":
		return true
	default:
		return false
	}
}

func requestURI(r *http.Request) string {
	if r.RequestURI != "" {
		return r.RequestURI
	}

	return r.URL.RequestURI()
}

type hashResponseWriter struct {
	http.ResponseWriter
	key        string
	buf        bytes.Buffer
	statusCode int
}

func (w *hashResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *hashResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	return w.buf.Write(p)
}

func (w *hashResponseWriter) flush() {
	body := w.buf.Bytes()

	mac := hmac.New(sha256.New, []byte(w.key))
	mac.Write(body)
	w.ResponseWriter.Header().Set("HashSHA256", hex.EncodeToString(mac.Sum(nil)))

	if w.statusCode != 0 {
		w.ResponseWriter.WriteHeader(w.statusCode)
	}

	if len(body) > 0 {
		_, _ = w.ResponseWriter.Write(body)
	}
}

func hashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if key == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			_ = r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if hashHeader := r.Header.Get("HashSHA256"); hashHeader != "" {
				mac := hmac.New(sha256.New, []byte(key))
				mac.Write(body)
				expected := hex.EncodeToString(mac.Sum(nil))
				if !hmac.Equal([]byte(hashHeader), []byte(expected)) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}

			cloned := r.Clone(r.Context())
			cloned.Body = io.NopCloser(bytes.NewReader(body))

			hrw := &hashResponseWriter{ResponseWriter: w, key: key}
			next.ServeHTTP(hrw, cloned)
			hrw.flush()
		})
	}
}
