package agent

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	models "github.com/lanefedov/metrics/internal/model"
)

func TestReporterReportSendsMetrics(t *testing.T) {
	var gotMethods []string
	var gotContentTypes []string
	var gotPaths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethods = append(gotMethods, r.Method)
		gotContentTypes = append(gotContentTypes, r.Header.Get("Content-Type"))
		gotPaths = append(gotPaths, r.URL.Path)
		_, _ = io.WriteString(w, "OK")
	}))
	defer server.Close()

	reporter := NewReporter(server.URL, server.Client())
	metrics := []Metric{
		{Name: "PollCount", Type: models.Counter, CounterValue: 7},
		{Name: "RandomValue", Type: models.Gauge, GaugeValue: 42.5},
	}

	if err := reporter.Report(metrics); err != nil {
		t.Fatalf("report metrics: %v", err)
	}

	wantMethods := []string{http.MethodPost, http.MethodPost}
	if !reflect.DeepEqual(gotMethods, wantMethods) {
		t.Fatalf("methods: got %v, want %v", gotMethods, wantMethods)
	}

	wantContentTypes := []string{"text/plain", "text/plain"}
	if !reflect.DeepEqual(gotContentTypes, wantContentTypes) {
		t.Fatalf("content types: got %v, want %v", gotContentTypes, wantContentTypes)
	}

	wantPaths := []string{
		"/update/counter/PollCount/7",
		"/update/gauge/RandomValue/42.5",
	}
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("paths: got %v, want %v", gotPaths, wantPaths)
	}
}

func TestReporterReportReturnsTransportError(t *testing.T) {
	reporter := NewReporter("http://localhost:8080", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, assertiveError("transport failed")
	}))

	err := reporter.Report([]Metric{
		{Name: "PollCount", Type: models.Counter, CounterValue: 1},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "transport failed" {
		t.Fatalf("error: got %q, want %q", err.Error(), "transport failed")
	}
}

func TestReporterReportReturnsStatusError(t *testing.T) {
	reporter := NewReporter("http://localhost:8080", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(http.NoBody),
		}, nil
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
}

func TestReporterAcceptsAddressWithoutScheme(t *testing.T) {
	var gotHost string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reporter := NewReporter(server.Listener.Addr().String(), server.Client())

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

func TestReporterReportJoinsMultipleErrors(t *testing.T) {
	firstErr := assertiveError("transport failed 1")
	secondErr := assertiveError("transport failed 2")
	callNumber := 0

	reporter := NewReporter("http://localhost:8080", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		callNumber++
		if callNumber == 1 {
			return nil, firstErr
		}
		return nil, secondErr
	}))

	err := reporter.Report([]Metric{
		{Name: "PollCount", Type: models.Counter, CounterValue: 1},
		{Name: "Alloc", Type: models.Gauge, GaugeValue: 1.5},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, firstErr) {
		t.Fatalf("expected joined error to include %q, got %v", firstErr, err)
	}
	if !errors.Is(err, secondErr) {
		t.Fatalf("expected joined error to include %q, got %v", secondErr, err)
	}
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

type assertiveError string

func (e assertiveError) Error() string {
	return string(e)
}
