package server

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestLoggingMiddlewareLogsRequestAndResponse(t *testing.T) {
	var output bytes.Buffer
	logger := zerolog.New(&output)
	handler := loggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/update/counter/test/1", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusCreated)
	}

	scanner := bufio.NewScanner(&output)
	var entries []map[string]any
	for scanner.Scan() {
		var entry map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("unmarshal log entry: %v", err)
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scan logs: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("log entries: got %d, want %d", len(entries), 2)
	}

	requestLog := entries[0]
	if got := requestLog["level"]; got != "info" {
		t.Fatalf("request log level: got %v, want %q", got, "info")
	}
	if got := requestLog["uri"]; got != "/update/counter/test/1" {
		t.Fatalf("request uri: got %v, want %q", got, "/update/counter/test/1")
	}
	if got := requestLog["method"]; got != http.MethodPost {
		t.Fatalf("request method: got %v, want %q", got, http.MethodPost)
	}
	if _, ok := requestLog["duration"]; !ok {
		t.Fatal("request duration is missing")
	}

	responseLog := entries[1]
	if got := responseLog["level"]; got != "info" {
		t.Fatalf("response log level: got %v, want %q", got, "info")
	}
	if got := responseLog["status"]; got != float64(http.StatusCreated) {
		t.Fatalf("response status: got %v, want %d", got, http.StatusCreated)
	}
	if got := responseLog["size"]; got != float64(len("hello")) {
		t.Fatalf("response size: got %v, want %d", got, len("hello"))
	}
}

func TestRequestDecompressionMiddlewareDecompressesGzipBody(t *testing.T) {
	const wantBody = `{"id":"Alloc","type":"gauge","value":123.45}`

	var gotBody string
	var gotEncoding string
	handler := requestDecompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		gotBody = string(body)
		gotEncoding = r.Header.Get("Content-Encoding")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(gzipBytes(t, wantBody)))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
	if gotBody != wantBody {
		t.Fatalf("body: got %q, want %q", gotBody, wantBody)
	}
	if gotEncoding != "" {
		t.Fatalf("content encoding: got %q, want empty", gotEncoding)
	}
}

func TestRequestDecompressionMiddlewareRejectsBrokenGzip(t *testing.T) {
	called := false
	handler := requestDecompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader("not a gzip stream"))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if called {
		t.Fatal("handler should not be called for broken gzip")
	}
}

func TestResponseCompressionMiddlewareCompressesEligibleResponses(t *testing.T) {
	testCases := []struct {
		name        string
		contentType string
		body        string
	}{
		{
			name:        "json",
			contentType: "application/json",
			body:        `{"status":"ok"}`,
		},
		{
			name:        "html",
			contentType: "text/html; charset=utf-8",
			body:        "<html><body>ok</body></html>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := responseCompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tc.contentType)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tc.body))
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
			}
			if got := rr.Header().Get("Content-Encoding"); got != "gzip" {
				t.Fatalf("content encoding: got %q, want %q", got, "gzip")
			}
			if got := rr.Header().Get("Vary"); got != "Accept-Encoding" {
				t.Fatalf("vary: got %q, want %q", got, "Accept-Encoding")
			}

			gotBody := gunzipString(t, rr.Body.Bytes())
			if gotBody != tc.body {
				t.Fatalf("body: got %q, want %q", gotBody, tc.body)
			}
		})
	}
}

func TestResponseCompressionMiddlewareSkipsUnsupportedContentType(t *testing.T) {
	const wantBody = "plain text response"

	handler := responseCompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(wantBody))
	}))

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("content encoding: got %q, want empty", got)
	}
	if got := rr.Body.String(); got != wantBody {
		t.Fatalf("body: got %q, want %q", got, wantBody)
	}
}

func gzipBytes(t *testing.T, value string) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write([]byte(value)); err != nil {
		t.Fatalf("write gzip body: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	return buffer.Bytes()
}

func gunzipString(t *testing.T, data []byte) string {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new gzip reader: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close gzip reader: %v", err)
		}
	}()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read gzip body: %v", err)
	}

	return string(body)
}
