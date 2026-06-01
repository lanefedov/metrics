package handler

import (
	"errors"
	"net/http"

	"github.com/lanefedov/metrics/internal/service"
)

// JSONUpdatesHandler обрабатывает POST /updates с JSON-массивом метрик.
type JSONUpdatesHandler struct {
	service *service.MetricsService
}

// NewJSONUpdatesHandler создаёт JSON-обработчик батчевого обновления метрик.
func NewJSONUpdatesHandler(metricsService *service.MetricsService) *JSONUpdatesHandler {
	return &JSONUpdatesHandler{service: metricsService}
}

// ServeHTTP реализует net/http.Handler.
func (h *JSONUpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !isJSONContentType(r.Header.Get("Content-Type")) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metrics, err := decodeMetricsJSON(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateMetrics(r.Context(), metrics); err != nil {
		if errors.Is(err, service.ErrInvalidMetric) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeMetricsJSON(w, http.StatusOK, metrics)
}
