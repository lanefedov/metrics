package storage

import "testing"

func TestMemStorage(t *testing.T) {
	s := NewMemStorage()

	s.AddCounter("c", 10)
	s.AddCounter("c", 5)
	s.SetGauge("g", 1.25)
	s.SetGauge("g", 2.5)

	if s.counters["c"] != 15 {
		t.Fatalf("counter: got %d, want 15", s.counters["c"])
	}
	if s.gauges["g"] != 2.5 {
		t.Fatalf("gauge: got %v, want 2.5", s.gauges["g"])
	}
}
