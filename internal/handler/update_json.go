package handler

import (
	"errors"
	"net/http"

	"github.com/lanefedov/metrics/internal/service"
)

// JSONUpdateHandler обрабатывает POST /update с JSON-телом.
type JSONUpdateHandler struct {
	service *service.MetricsService
}

// NewJSONUpdateHandler создаёт JSON-обработчик обновления метрик.
func NewJSONUpdateHandler(metricsService *service.MetricsService) *JSONUpdateHandler {
	return &JSONUpdateHandler{service: metricsService}
}

// ServeHTTP реализует net/http.Handler.
func (h *JSONUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.UpdateMetric(r.Context(), metric); err != nil {
		if errors.Is(err, service.ErrInvalidMetric) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeMetricJSON(w, http.StatusOK, metric)
}
