package handler

import (
	"context"
	"net/http"
	"time"
)

const databasePingTimeout = time.Second

// PingFunc проверяет доступность внешней зависимости.
type PingFunc func(context.Context) error

// PingHandler обрабатывает GET /ping.
type PingHandler struct {
	ping PingFunc
}

// NewPingHandler создаёт обработчик проверки подключения к базе данных.
func NewPingHandler(ping PingFunc) *PingHandler {
	return &PingHandler{ping: ping}
}

// ServeHTTP реализует net/http.Handler.
func (h *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.ping == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), databasePingTimeout)
	defer cancel()

	if err := h.ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
