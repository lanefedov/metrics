package server

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/lanefedov/metrics/internal/handler"
	"github.com/lanefedov/metrics/internal/service"
	"github.com/rs/zerolog"
)

// NewHandler создаёт дерево HTTP-обработчиков сервера метрик.
func NewHandler(metricsService *service.MetricsService, key string, databasePing ...func(context.Context) error) http.Handler {
	r := chi.NewRouter()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	var ping handler.PingFunc
	if len(databasePing) > 0 {
		ping = databasePing[0]
	}

	r.Use(loggingMiddleware(logger))
	r.Use(requestDecompressionMiddleware)
	r.Use(responseCompressionMiddleware)
	r.Use(hashMiddleware(key))
	r.Get("/ping", handler.NewPingHandler(ping).ServeHTTP)
	r.Post("/update", handler.NewJSONUpdateHandler(metricsService).ServeHTTP)
	r.Post("/update/", handler.NewJSONUpdateHandler(metricsService).ServeHTTP)
	r.Post("/updates", handler.NewJSONUpdatesHandler(metricsService).ServeHTTP)
	r.Post("/updates/", handler.NewJSONUpdatesHandler(metricsService).ServeHTTP)
	r.Post("/update/{type}/{name}/{value}", handler.NewUpdateHandler(metricsService).ServeHTTP)
	r.Post("/value", handler.NewJSONValueHandler(metricsService).ServeHTTP)
	r.Post("/value/", handler.NewJSONValueHandler(metricsService).ServeHTTP)
	r.Get("/value/{type}/{name}", handler.NewValueHandler(metricsService).ServeHTTP)
	r.Get("/", handler.NewListHandler(metricsService).ServeHTTP)

	return r
}
