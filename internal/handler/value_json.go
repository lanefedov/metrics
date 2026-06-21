package handler

import (
	"errors"
	"net/http"

	"github.com/lanefedov/metrics/internal/service"
)

// JSONValueHandler обрабатывает POST /value с JSON-телом.
type JSONValueHandler struct {
	service *service.MetricsService
}

// NewJSONValueHandler создаёт JSON-обработчик чтения метрики.
func NewJSONValueHandler(metricsService *service.MetricsService) *JSONValueHandler {
	return &JSONValueHandler{service: metricsService}
}

// ServeHTTP реализует net/http.Handler.
func (h *JSONValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !isJSONContentType(r.Header.Get("Content-Type")) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := decodeMetricJSON(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	responseMetric, err := h.service.GetMetric(r.Context(), metric)
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

	writeMetricJSON(w, http.StatusOK, *responseMetric)
}
