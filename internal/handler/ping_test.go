package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingHandlerReturnsOKWhenPingSucceeds(t *testing.T) {
	handler := NewPingHandler(func(context.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestPingHandlerReturnsInternalServerErrorWhenPingFails(t *testing.T) {
	handler := NewPingHandler(func(context.Context) error {
		return errors.New("database unavailable")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestPingHandlerReturnsInternalServerErrorWithoutPingFunc(t *testing.T) {
	handler := NewPingHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestPingHandlerRejectsNonGetRequests(t *testing.T) {
	handler := NewPingHandler(func(context.Context) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/ping", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}
