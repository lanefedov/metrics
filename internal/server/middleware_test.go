package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
