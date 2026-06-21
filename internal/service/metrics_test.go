package service

import (
	"context"
	"errors"
	"testing"

	models "github.com/lanefedov/metrics/internal/model"
	"github.com/lanefedov/metrics/internal/storage"
)

func TestMetricsServiceUpdateMetrics(t *testing.T) {
	store := storage.NewMemStorage()
	afterUpdateCalls := 0
	svc := NewMetricsServiceWithAfterUpdate(store, func() error {
		afterUpdateCalls++
		return nil
	})

	firstDelta := int64(10)
	secondDelta := int64(5)
	firstGauge := 1.25
	secondGauge := 2.5

	err := svc.UpdateMetrics(context.Background(), []models.Metrics{
		{ID: "c", MType: models.Counter, Delta: &firstDelta},
		{ID: "g", MType: models.Gauge, Value: &firstGauge},
		{ID: "c", MType: models.Counter, Delta: &secondDelta},
		{ID: "g", MType: models.Gauge, Value: &secondGauge},
	})
	if err != nil {
		t.Fatalf("update metrics: %v", err)
	}
	if afterUpdateCalls != 1 {
		t.Fatalf("after update calls: got %d, want 1", afterUpdateCalls)
	}

	counter, ok, err := store.GetCounter(context.Background(), "c")
	if err != nil {
		t.Fatalf("get counter: %v", err)
	}
	if !ok || counter != 15 {
		t.Fatalf("counter: got (%d, %t), want (15, true)", counter, ok)
	}

	gauge, ok, err := store.GetGauge(context.Background(), "g")
	if err != nil {
		t.Fatalf("get gauge: %v", err)
	}
	if !ok || gauge != 2.5 {
		t.Fatalf("gauge: got (%v, %t), want (2.5, true)", gauge, ok)
	}
}

func TestMetricsServiceUpdateMetricsValidatesBeforeWriting(t *testing.T) {
	store := storage.NewMemStorage()
	afterUpdateCalls := 0
	svc := NewMetricsServiceWithAfterUpdate(store, func() error {
		afterUpdateCalls++
		return nil
	})

	value := 1.25
	err := svc.UpdateMetrics(context.Background(), []models.Metrics{
		{ID: "g", MType: models.Gauge, Value: &value},
		{ID: "broken", MType: models.Counter},
	})
	if !errors.Is(err, ErrInvalidMetric) {
		t.Fatalf("error: got %v, want ErrInvalidMetric", err)
	}
	if afterUpdateCalls != 0 {
		t.Fatalf("after update calls: got %d, want 0", afterUpdateCalls)
	}
	if _, ok, err := store.GetGauge(context.Background(), "g"); err != nil || ok {
		t.Fatalf("gauge should not be written, got ok=%t err=%v", ok, err)
	}
}

func TestMetricsServiceUpdateMetricsSkipsEmptyBatch(t *testing.T) {
	store := storage.NewMemStorage()
	afterUpdateCalls := 0
	svc := NewMetricsServiceWithAfterUpdate(store, func() error {
		afterUpdateCalls++
		return nil
	})

	if err := svc.UpdateMetrics(context.Background(), nil); err != nil {
		t.Fatalf("update empty batch: %v", err)
	}
	if afterUpdateCalls != 0 {
		t.Fatalf("after update calls: got %d, want 0", afterUpdateCalls)
	}
}
