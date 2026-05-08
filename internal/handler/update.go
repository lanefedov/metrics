package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	models "github.com/lanefedov/metrics/internal/model"
	"github.com/lanefedov/metrics/internal/storage"
)

// UpdateHandler обрабатывает POST /update/{type}/{name}/{value}.
type UpdateHandler struct {
	store storage.MetricsUpdater
}

// NewUpdateHandler создаёт обработчик обновления метрик.
func NewUpdateHandler(store storage.MetricsUpdater) *UpdateHandler {
	return &UpdateHandler{store: store}
}

// ServeHTTP реализует net/http.Handler.
func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if ct := r.Header.Get("Content-Type"); ct != "" {
		base, _, _ := strings.Cut(ct, ";")
		if strings.TrimSpace(strings.ToLower(base)) != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	valueStr := chi.URLParam(r, "value")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	switch metricType {
	case models.Counter:
		v, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.store.AddCounter(name, v)
	case models.Gauge:
		v, err := strconv.ParseFloat(valueStr, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.store.SetGauge(name, v)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
