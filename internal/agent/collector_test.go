package agent

import (
	"runtime"
	"testing"

	models "github.com/lanefedov/metrics/internal/model"
)

func TestCollectorCollectsAllRequiredMetrics(t *testing.T) {
	reader := fakeStatsReader{
		stats: runtime.MemStats{
			Alloc:         1,
			BuckHashSys:   2,
			Frees:         3,
			GCCPUFraction: 4.5,
			GCSys:         5,
			HeapAlloc:     6,
			HeapIdle:      7,
			HeapInuse:     8,
			HeapObjects:   9,
			HeapReleased:  10,
			HeapSys:       11,
			LastGC:        12,
			Lookups:       13,
			MCacheInuse:   14,
			MCacheSys:     15,
			MSpanInuse:    16,
			MSpanSys:      17,
			Mallocs:       18,
			NextGC:        19,
			NumForcedGC:   20,
			NumGC:         21,
			OtherSys:      22,
			PauseTotalNs:  23,
			StackInuse:    24,
			StackSys:      25,
			Sys:           26,
			TotalAlloc:    27,
		},
	}
	collector := newCollector(reader, func() float64 { return 99.9 })

	collector.Collect()

	got := collector.Snapshot()
	if len(got) != 29 {
		t.Fatalf("metric count: got %d, want 29", len(got))
	}

	metricsByName := make(map[string]Metric, len(got))
	for _, metric := range got {
		metricsByName[metric.Name] = metric
	}

	requiredGaugeNames := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", randomValueMetric,
	}
	for _, name := range requiredGaugeNames {
		metric, ok := metricsByName[name]
		if !ok {
			t.Fatalf("missing metric %q", name)
		}
		if metric.Type != models.Gauge {
			t.Fatalf("metric type for %q: got %q, want %q", name, metric.Type, models.Gauge)
		}
	}

	pollCount, ok := metricsByName[pollCountMetric]
	if !ok {
		t.Fatalf("missing metric %q", pollCountMetric)
	}
	if pollCount.Type != models.Counter {
		t.Fatalf("poll count type: got %q, want %q", pollCount.Type, models.Counter)
	}
	if pollCount.CounterValue != 1 {
		t.Fatalf("poll count value: got %d, want 1", pollCount.CounterValue)
	}
}

func TestCollectorIncrementsPollCountAndUpdatesRandomValue(t *testing.T) {
	reader := fakeStatsReader{}
	randomValues := []float64{1.25, 2.5}
	nextRandom := 0
	collector := newCollector(reader, func() float64 {
		value := randomValues[nextRandom]
		nextRandom++
		return value
	})

	collector.Collect()
	first := snapshotByName(collector.Snapshot())
	collector.Collect()
	second := snapshotByName(collector.Snapshot())

	if first[pollCountMetric].CounterValue != 1 {
		t.Fatalf("first poll count: got %d, want 1", first[pollCountMetric].CounterValue)
	}
	if second[pollCountMetric].CounterValue != 2 {
		t.Fatalf("second poll count: got %d, want 2", second[pollCountMetric].CounterValue)
	}
	if first[randomValueMetric].GaugeValue != 1.25 {
		t.Fatalf("first random value: got %v, want %v", first[randomValueMetric].GaugeValue, 1.25)
	}
	if second[randomValueMetric].GaugeValue != 2.5 {
		t.Fatalf("second random value: got %v, want %v", second[randomValueMetric].GaugeValue, 2.5)
	}
}

type fakeStatsReader struct {
	stats runtime.MemStats
}

func (f fakeStatsReader) Read(stats *runtime.MemStats) {
	*stats = f.stats
}

func snapshotByName(metrics []Metric) map[string]Metric {
	result := make(map[string]Metric, len(metrics))
	for _, metric := range metrics {
		result[metric.Name] = metric
	}
	return result
}
