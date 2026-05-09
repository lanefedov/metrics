package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	models "github.com/lanefedov/metrics/internal/model"
	"github.com/lanefedov/metrics/internal/service"
)

// ValueHandler обрабатывает GET /value/{type}/{name}.
type ValueHandler struct {
	service *service.MetricsService
}

// NewValueHandler создаёт обработчик чтения метрики.
func NewValueHandler(metricsService *service.MetricsService) *ValueHandler {
	return &ValueHandler{service: metricsService}
}

// ServeHTTP реализует net/http.Handler.
func (h *ValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric := models.Metrics{
		ID:    chi.URLParam(r, "name"),
		MType: chi.URLParam(r, "type"),
	}
	if metric.ID == "" {
		http.NotFound(w, r)
		return
	}

	responseMetric, err := h.service.GetMetric(metric)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidMetric):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, service.ErrMetricNotFound):
			http.NotFound(w, r)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	var value string
	switch responseMetric.MType {
	case models.Gauge:
		value = strconv.FormatFloat(*responseMetric.Value, 'f', -1, 64)
	case models.Counter:
		value = strconv.FormatInt(*responseMetric.Delta, 10)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(value))
}
