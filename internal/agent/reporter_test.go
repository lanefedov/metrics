package agent

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	models "github.com/lanefedov/metrics/internal/model"
)

func TestReporterReportSendsMetrics(t *testing.T) {
	var gotMethod string
	var gotContentType string
	var gotContentEncoding string
	var gotPath string
	var gotBody []models.Metrics
	var gotCalls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCalls++
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotContentEncoding = r.Header.Get("Content-Encoding")
		gotPath = r.URL.Path

		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("new gzip reader: %v", err)
		}

		if err := json.NewDecoder(reader).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if err := reader.Close(); err != nil {
			t.Fatalf("close gzip reader: %v", err)
		}

		_, _ = io.WriteString(w, "OK")
	}))
	defer server.Close()

	reporter := NewReporter(server.URL, "", resty.NewWithClient(server.Client()))
	metrics := []Metric{
		{Name: "PollCount", Type: models.Counter, CounterValue: 7},
		{Name: "RandomValue", Type: models.Gauge, GaugeValue: 42.5},
	}

	if err := reporter.Report(metrics); err != nil {
		t.Fatalf("report metrics: %v", err)
	}

	if gotCalls != 1 {
		t.Fatalf("calls: got %d, want 1", gotCalls)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method: got %q, want %q", gotMethod, http.MethodPost)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content type: got %q, want %q", gotContentType, "application/json")
	}
	if gotContentEncoding != "gzip" {
		t.Fatalf("content encoding: got %q, want %q", gotContentEncoding, "gzip")
	}
	if gotPath != "/updates/" {
		t.Fatalf("path: got %q, want %q", gotPath, "/updates/")
	}

	wantBody := []models.Metrics{
		{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: int64Ptr(7),
		},
		{
			ID:    "RandomValue",
			MType: models.Gauge,
			Value: float64Ptr(42.5),
		},
	}
	if !reflect.DeepEqual(gotBody, wantBody) {
		t.Fatalf("body: got %#v, want %#v", gotBody, wantBody)
	}
}

func TestReporterReportSkipsEmptyBatch(t *testing.T) {
	called := false
	reporter := NewReporter("http://localhost:8080", "", resty.NewWithClient(&http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			called = true
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(http.NoBody),
				Header:     make(http.Header),
			}, nil
		}),
	}))

	if err := reporter.Report(nil); err != nil {
		t.Fatalf("report empty batch: %v", err)
	}
	if called {
		t.Fatal("empty batch should not trigger HTTP request")
	}
}

func TestReporterReportReturnsTransportError(t *testing.T) {
	transportErr := assertiveError("transport failed")
	var calls int
	reporter := NewReporter("http://localhost:8080", "", resty.NewWithClient(&http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			calls++
			return nil, transportErr
		}),
	}))
	disableReporterRetrySleep(reporter)

	err := reporter.Report([]Metric{
		{Name: "PollCount", Type: models.Counter, CounterValue: 1},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, transportErr) {
		t.Fatalf("expected error to include %q, got %v", transportErr, err)
	}
	if calls != 4 {
		t.Fatalf("calls: got %d, want 4", calls)
	}
}

func TestReporterReportRetriesTransportErrorUntilSuccess(t *testing.T) {
	transportErr := assertiveError("transport failed")
	var calls int
	reporter := NewReporter("http://localhost:8080", "", resty.NewWithClient(&http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			calls++
			if calls < 3 {
				return nil, transportErr
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(http.NoBody),
				Header:     make(http.Header),
			}, nil
		}),
	}))
	disableReporterRetrySleep(reporter)

	err := reporter.Report([]Metric{
		{Name: "PollCount", Type: models.Counter, CounterValue: 1},
	})
	if err != nil {
		t.Fatalf("report metrics: %v", err)
	}
	if calls != 3 {
		t.Fatalf("calls: got %d, want 3", calls)
	}
}

func TestReporterReportReturnsStatusError(t *testing.T) {
	var calls int
	reporter := NewReporter("http://localhost:8080", "", resty.NewWithClient(&http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			calls++
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(http.NoBody),
				Header:     make(http.Header),
			}, nil
		}),
	}))

	err := reporter.Report([]Metric{
		{Name: "Alloc", Type: models.Gauge, GaugeValue: 1.5},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "unexpected status code: 400" {
		t.Fatalf("error: got %q, want %q", err.Error(), "unexpected status code: 400")
	}
	if calls != 1 {
		t.Fatalf("calls: got %d, want 1", calls)
	}
}

func TestReporterAcceptsAddressWithoutScheme(t *testing.T) {
	var gotHost string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reporter := NewReporter(server.Listener.Addr().String(), "", resty.NewWithClient(server.Client()))

	err := reporter.Report([]Metric{
		{Name: "Alloc", Type: models.Gauge, GaugeValue: 1.5},
	})
	if err != nil {
		t.Fatalf("report metrics: %v", err)
	}
	if gotHost != server.Listener.Addr().String() {
		t.Fatalf("host: got %q, want %q", gotHost, server.Listener.Addr().String())
	}
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type assertiveError string

func (e assertiveError) Error() string {
	return string(e)
}

func disableReporterRetrySleep(reporter *Reporter) {
	reporter.retrySleep = func(context.Context, time.Duration) error {
		return nil
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}
