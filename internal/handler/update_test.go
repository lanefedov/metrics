package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	models "github.com/lanefedov/metrics/internal/model"
	"github.com/lanefedov/metrics/internal/service"
)

func TestUpdateHandler(t *testing.T) {
	store := &fakeMetricsStorage{}
	h := newTestRouter(store)

	tests := []struct {
		name       string
		method     string
		path       string
		headerCT   string
		wantStatus int
	}{
		{
			name:       "counter ok",
			method:     http.MethodPost,
			path:       "/update/counter/someMetric/527",
			headerCT:   "text/plain",
			wantStatus: http.StatusOK,
		},
		{
			name:       "gauge ok",
			method:     http.MethodPost,
			path:       "/update/gauge/x/1.5",
			headerCT:   "text/plain",
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty metric name",
			method:     http.MethodPost,
			path:       "/update/counter//527",
			headerCT:   "text/plain",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "bad metric type",
			method:     http.MethodPost,
			path:       "/update/unknown/x/1",
			headerCT:   "text/plain",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad counter value",
			method:     http.MethodPost,
			path:       "/update/counter/x/3.14",
			headerCT:   "text/plain",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad gauge value",
			method:     http.MethodPost,
			path:       "/update/gauge/x/not_a_float",
			headerCT:   "text/plain",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			path:       "/update/counter/x/1",
			headerCT:   "text/plain",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "wrong content type",
			method:     http.MethodPost,
			path:       "/update/counter/x/1",
			headerCT:   "application/json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no content type",
			method:     http.MethodPost,
			path:       "/update/counter/x/1",
			headerCT:   "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "text plain with charset",
			method:     http.MethodPost,
			path:       "/update/counter/x/2",
			headerCT:   "text/plain; charset=utf-8",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.headerCT != "" {
				req.Header.Set("Content-Type", tt.headerCT)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestUpdateHandlerStoresGauge(t *testing.T) {
	store := &fakeMetricsStorage{}
	h := newTestRouter(store)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/42.5", nil)
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
	if store.lastGaugeName != "Alloc" {
		t.Fatalf("gauge name: got %q, want %q", store.lastGaugeName, "Alloc")
	}
	if store.lastGaugeValue != 42.5 {
		t.Fatalf("gauge value: got %v, want %v", store.lastGaugeValue, 42.5)
	}
}

func TestUpdateHandlerStoresCounter(t *testing.T) {
	store := &fakeMetricsStorage{}
	h := newTestRouter(store)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/7", nil)
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
	if store.lastCounterName != "PollCount" {
		t.Fatalf("counter name: got %q, want %q", store.lastCounterName, "PollCount")
	}
	if store.lastCounterDelta != 7 {
		t.Fatalf("counter delta: got %d, want %d", store.lastCounterDelta, 7)
	}
}

func TestValueHandler(t *testing.T) {
	store := &fakeMetricsStorage{}
	_ = store.SetGauge(context.Background(), "Alloc", 42.5)
	_ = store.AddCounter(context.Background(), "PollCount", 7)

	h := newTestRouter(store)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "counter ok",
			path:       "/value/counter/PollCount",
			wantStatus: http.StatusOK,
			wantBody:   "7",
		},
		{
			name:       "gauge ok",
			path:       "/value/gauge/Alloc",
			wantStatus: http.StatusOK,
			wantBody:   "42.5",
		},
		{
			name:       "unknown metric",
			path:       "/value/counter/Unknown",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "bad metric type",
			path:       "/value/unknown/Alloc",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && strings.TrimSpace(rr.Body.String()) != tt.wantBody {
				t.Fatalf("body: got %q, want %q", rr.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestJSONUpdateHandler(t *testing.T) {
	store := &fakeMetricsStorage{}
	h := newTestRouter(store)

	gaugeValue := 42.5
	counterDelta := int64(7)

	tests := []struct {
		name             string
		contentType      string
		body             string
		wantStatus       int
		wantContentType  string
		wantResponseBody models.Metrics
	}{
		{
			name:            "gauge ok",
			contentType:     "application/json",
			body:            mustMetricJSON(t, models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &gaugeValue}),
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantResponseBody: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: &gaugeValue,
			},
		},
		{
			name:            "counter ok",
			contentType:     "application/json; charset=utf-8",
			body:            mustMetricJSON(t, models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &counterDelta}),
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantResponseBody: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: &counterDelta,
			},
		},
		{
			name:        "wrong content type",
			contentType: "text/plain",
			body:        mustMetricJSON(t, models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &gaugeValue}),
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid json",
			contentType: "application/json",
			body:        `{"id":"Alloc",`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "missing gauge value",
			contentType: "application/json",
			body:        mustMetricJSON(t, models.Metrics{ID: "Alloc", MType: models.Gauge}),
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "unknown metric type",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"unknown","value":42.5}`,
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantContentType != "" {
				if got := rr.Header().Get("Content-Type"); got != tt.wantContentType {
					t.Fatalf("content type: got %q, want %q", got, tt.wantContentType)
				}

				var gotBody models.Metrics
				if err := json.NewDecoder(rr.Body).Decode(&gotBody); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if !equalMetrics(gotBody, tt.wantResponseBody) {
					t.Fatalf("response body: got %+v, want %+v", gotBody, tt.wantResponseBody)
				}
			}
		})
	}

	if store.lastGaugeName != "Alloc" || store.lastGaugeValue != 42.5 {
		t.Fatalf("stored gauge: got %q=%v, want %q=%v", store.lastGaugeName, store.lastGaugeValue, "Alloc", 42.5)
	}
	if store.lastCounterName != "PollCount" || store.lastCounterDelta != 7 {
		t.Fatalf("stored counter: got %q=%d, want %q=%d", store.lastCounterName, store.lastCounterDelta, "PollCount", 7)
	}
}

func TestJSONValueHandler(t *testing.T) {
	store := &fakeMetricsStorage{}
	_ = store.SetGauge(context.Background(), "Alloc", 42.5)
	_ = store.AddCounter(context.Background(), "PollCount", 7)

	h := newTestRouter(store)

	tests := []struct {
		name             string
		contentType      string
		body             string
		wantStatus       int
		wantContentType  string
		wantResponseBody models.Metrics
	}{
		{
			name:            "gauge ok",
			contentType:     "application/json",
			body:            `{"id":"Alloc","type":"gauge"}`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantResponseBody: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: float64Ptr(42.5),
			},
		},
		{
			name:            "counter ok",
			contentType:     "application/json",
			body:            `{"id":"PollCount","type":"counter"}`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantResponseBody: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: int64Ptr(7),
			},
		},
		{
			name:        "unknown metric",
			contentType: "application/json",
			body:        `{"id":"Unknown","type":"counter"}`,
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "bad metric type",
			contentType: "application/json",
			body:        `{"id":"Alloc","type":"unknown"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "wrong content type",
			contentType: "text/plain",
			body:        `{"id":"Alloc","type":"gauge"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid json",
			contentType: "application/json",
			body:        `{"id":"Alloc"`,
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantContentType != "" {
				if got := rr.Header().Get("Content-Type"); got != tt.wantContentType {
					t.Fatalf("content type: got %q, want %q", got, tt.wantContentType)
				}

				var gotBody models.Metrics
				if err := json.NewDecoder(rr.Body).Decode(&gotBody); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if !equalMetrics(gotBody, tt.wantResponseBody) {
					t.Fatalf("response body: got %+v, want %+v", gotBody, tt.wantResponseBody)
				}
			}
		})
	}
}

func TestListHandler(t *testing.T) {
	store := &fakeMetricsStorage{}
	_ = store.SetGauge(context.Background(), "Alloc", 42.5)
	_ = store.AddCounter(context.Background(), "PollCount", 7)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	newTestRouter(store).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	for _, want := range []string{
		"<html>",
		"counter PollCount: 7",
		"gauge Alloc: 42.5",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body %q does not contain %q", body, want)
		}
	}
}

func newTestRouter(store *fakeMetricsStorage) http.Handler {
	metricsService := service.NewMetricsService(store)
	r := chi.NewRouter()
	r.Post("/update", NewJSONUpdateHandler(metricsService).ServeHTTP)
	r.Post("/update/", NewJSONUpdateHandler(metricsService).ServeHTTP)
	r.Post("/update/{type}/{name}/{value}", NewUpdateHandler(metricsService).ServeHTTP)
	r.Post("/value", NewJSONValueHandler(metricsService).ServeHTTP)
	r.Post("/value/", NewJSONValueHandler(metricsService).ServeHTTP)
	r.Get("/value/{type}/{name}", NewValueHandler(metricsService).ServeHTTP)
	r.Get("/", NewListHandler(metricsService).ServeHTTP)

	return r
}

func mustMetricJSON(t *testing.T, metric models.Metrics) string {
	t.Helper()

	body, err := json.Marshal(metric)
	if err != nil {
		t.Fatalf("marshal metric: %v", err)
	}

	return string(body)
}

func equalMetrics(got, want models.Metrics) bool {
	if got.ID != want.ID || got.MType != want.MType || got.Hash != want.Hash {
		return false
	}
	if !equalInt64Ptr(got.Delta, want.Delta) {
		return false
	}
	if !equalFloat64Ptr(got.Value, want.Value) {
		return false
	}

	return true
}

func equalInt64Ptr(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}

func equalFloat64Ptr(a, b *float64) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}

func int64Ptr(v int64) *int64 {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}

type fakeMetricsStorage struct {
	gauges           map[string]float64
	counters         map[string]int64
	lastGaugeName    string
	lastGaugeValue   float64
	lastCounterName  string
	lastCounterDelta int64
}

func (f *fakeMetricsStorage) SetGauge(_ context.Context, name string, value float64) error {
	if f.gauges == nil {
		f.gauges = make(map[string]float64)
	}
	f.gauges[name] = value
	f.lastGaugeName = name
	f.lastGaugeValue = value
	return nil
}

func (f *fakeMetricsStorage) AddCounter(_ context.Context, name string, delta int64) error {
	if f.counters == nil {
		f.counters = make(map[string]int64)
	}
	f.counters[name] += delta
	f.lastCounterName = name
	f.lastCounterDelta = delta
	return nil
}

func (f *fakeMetricsStorage) GetGauge(_ context.Context, name string) (float64, bool, error) {
	value, ok := f.gauges[name]
	return value, ok, nil
}

func (f *fakeMetricsStorage) GetCounter(_ context.Context, name string) (int64, bool, error) {
	value, ok := f.counters[name]
	return value, ok, nil
}

func (f *fakeMetricsStorage) ListGauges(_ context.Context) (map[string]float64, error) {
	gauges := make(map[string]float64, len(f.gauges))
	for name, value := range f.gauges {
		gauges[name] = value
	}

	return gauges, nil
}

func (f *fakeMetricsStorage) ListCounters(_ context.Context) (map[string]int64, error) {
	counters := make(map[string]int64, len(f.counters))
	for name, value := range f.counters {
		counters[name] = value
	}

	return counters, nil
}
