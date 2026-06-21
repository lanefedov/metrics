package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lanefedov/metrics/internal/service"
	"github.com/lanefedov/metrics/internal/storage"
)

func TestNewHandlerRegistersPingRoute(t *testing.T) {
	metricsService := service.NewMetricsService(storage.NewMemStorage())
	handler := NewHandler(metricsService, "", func(context.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
}
