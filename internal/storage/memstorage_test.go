package storage

import (
	"context"
	"testing"

	models "github.com/lanefedov/metrics/internal/model"
)

func TestMemStorage(t *testing.T) {
	s := NewMemStorage()

	ctx := context.Background()
	_ = s.AddCounter(ctx, "c", 10)
	_ = s.AddCounter(ctx, "c", 5)
	_ = s.SetGauge(ctx, "g", 1.25)
	_ = s.SetGauge(ctx, "g", 2.5)

	if s.counters["c"] != 15 {
		t.Fatalf("counter: got %d, want 15", s.counters["c"])
	}
	if s.gauges["g"] != 2.5 {
		t.Fatalf("gauge: got %v, want 2.5", s.gauges["g"])
	}
}

func TestMemStorageUpdateMetrics(t *testing.T) {
	s := NewMemStorage()
	ctx := context.Background()

	firstDelta := int64(10)
	secondDelta := int64(5)
	firstGauge := 1.25
	secondGauge := 2.5

	err := s.UpdateMetrics(ctx, []models.Metrics{
		{ID: "c", MType: models.Counter, Delta: &firstDelta},
		{ID: "g", MType: models.Gauge, Value: &firstGauge},
		{ID: "c", MType: models.Counter, Delta: &secondDelta},
		{ID: "g", MType: models.Gauge, Value: &secondGauge},
	})
	if err != nil {
		t.Fatalf("update metrics: %v", err)
	}

	counter, ok, err := s.GetCounter(ctx, "c")
	if err != nil {
		t.Fatalf("get counter: %v", err)
	}
	if !ok || counter != 15 {
		t.Fatalf("counter: got (%d, %t), want (15, true)", counter, ok)
	}

	gauge, ok, err := s.GetGauge(ctx, "g")
	if err != nil {
		t.Fatalf("get gauge: %v", err)
	}
	if !ok || gauge != 2.5 {
		t.Fatalf("gauge: got (%v, %t), want (2.5, true)", gauge, ok)
	}
}
