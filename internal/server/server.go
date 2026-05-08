package server

import (
	"net/http"

	"github.com/lanefedov/metrics/internal/handler"
	"github.com/lanefedov/metrics/internal/storage"
)

// NewHandler builds the HTTP handler tree for the metrics server.
func NewHandler(store storage.MetricsStorage) http.Handler {
	return handler.NewUpdateHandler(store)
}
