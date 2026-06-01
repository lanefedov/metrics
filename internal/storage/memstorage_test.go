package storage

import (
	"context"
	"testing"
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
