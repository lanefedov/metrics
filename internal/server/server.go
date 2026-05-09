package server

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/lanefedov/metrics/internal/handler"
	"github.com/lanefedov/metrics/internal/service"
	"github.com/lanefedov/metrics/internal/storage"
	"github.com/rs/zerolog"
)

// NewHandler создаёт дерево HTTP-обработчиков сервера метрик.
func NewHandler(store storage.MetricsStorage) http.Handler {
	r := chi.NewRouter()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	metricsService := service.NewMetricsService(store)

	r.Use(loggingMiddleware(logger))
	r.Use(requestDecompressionMiddleware)
	r.Use(responseCompressionMiddleware)
	r.Post("/update", handler.NewJSONUpdateHandler(metricsService).ServeHTTP)
	r.Post("/update/", handler.NewJSONUpdateHandler(metricsService).ServeHTTP)
	r.Post("/update/{type}/{name}/{value}", handler.NewUpdateHandler(metricsService).ServeHTTP)
	r.Post("/value", handler.NewJSONValueHandler(metricsService).ServeHTTP)
	r.Post("/value/", handler.NewJSONValueHandler(metricsService).ServeHTTP)
	r.Get("/value/{type}/{name}", handler.NewValueHandler(metricsService).ServeHTTP)
	r.Get("/", handler.NewListHandler(metricsService).ServeHTTP)

	return r
}
