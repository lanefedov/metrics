package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateHandler(t *testing.T) {
	store := &fakeMetricsStorage{}
	h := NewUpdateHandler(store)

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
	h := NewUpdateHandler(store)

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
	h := NewUpdateHandler(store)

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

type fakeMetricsStorage struct {
	lastGaugeName    string
	lastGaugeValue   float64
	lastCounterName  string
	lastCounterDelta int64
}

func (f *fakeMetricsStorage) SetGauge(name string, value float64) {
	f.lastGaugeName = name
	f.lastGaugeValue = value
}

func (f *fakeMetricsStorage) AddCounter(name string, delta int64) {
	f.lastCounterName = name
	f.lastCounterDelta = delta
}
