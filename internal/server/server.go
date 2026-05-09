package server

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/lanefedov/metrics/internal/handler"
	"github.com/lanefedov/metrics/internal/storage"
	"github.com/rs/zerolog"
)

// NewHandler создаёт дерево HTTP-обработчиков сервера метрик.
func NewHandler(store storage.MetricsStorage) http.Handler {
	r := chi.NewRouter()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	r.Use(loggingMiddleware(logger))
	r.Post("/update/{type}/{name}/{value}", handler.NewUpdateHandler(store).ServeHTTP)
	r.Get("/value/{type}/{name}", handler.NewValueHandler(store).ServeHTTP)
	r.Get("/", handler.NewListHandler(store).ServeHTTP)

	return r
}
