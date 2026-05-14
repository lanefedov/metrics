package agent

import (
	"math/rand"
	"runtime"
	"sort"
	"sync"

	models "github.com/lanefedov/metrics/internal/model"
)

const (
	pollCountMetric   = "PollCount"
	randomValueMetric = "RandomValue"
)

type statsReader interface {
	Read(*runtime.MemStats)
}

type runtimeStatsReader struct{}

func (runtimeStatsReader) Read(stats *runtime.MemStats) {
	runtime.ReadMemStats(stats)
}

// Collector хранит в памяти последний снимок runtime-метрик.
type Collector struct {
	reader     statsReader
	randomFunc func() float64

	mu        sync.RWMutex
	pollCount int64
	metrics   map[string]Metric
}

// NewCollector создаёт сборщик runtime-метрик.
func NewCollector(reader statsReader, randomFunc func() float64) *Collector {
	if reader == nil {
		reader = runtimeStatsReader{}
	}
	if randomFunc == nil {
		randomFunc = rand.Float64
	}

	return &Collector{
		reader:     reader,
		randomFunc: randomFunc,
		metrics:    make(map[string]Metric),
	}
}

// Collect обновляет текущий снимок метрик.
func (c *Collector) Collect() {
	var stats runtime.MemStats
	c.reader.Read(&stats)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.pollCount++

	c.setGauge("Alloc", float64(stats.Alloc))
	c.setGauge("BuckHashSys", float64(stats.BuckHashSys))
	c.setGauge("Frees", float64(stats.Frees))
	c.setGauge("GCCPUFraction", stats.GCCPUFraction)
	c.setGauge("GCSys", float64(stats.GCSys))
	c.setGauge("HeapAlloc", float64(stats.HeapAlloc))
	c.setGauge("HeapIdle", float64(stats.HeapIdle))
	c.setGauge("HeapInuse", float64(stats.HeapInuse))
	c.setGauge("HeapObjects", float64(stats.HeapObjects))
	c.setGauge("HeapReleased", float64(stats.HeapReleased))
	c.setGauge("HeapSys", float64(stats.HeapSys))
	c.setGauge("LastGC", float64(stats.LastGC))
	c.setGauge("Lookups", float64(stats.Lookups))
	c.setGauge("MCacheInuse", float64(stats.MCacheInuse))
	c.setGauge("MCacheSys", float64(stats.MCacheSys))
	c.setGauge("MSpanInuse", float64(stats.MSpanInuse))
	c.setGauge("MSpanSys", float64(stats.MSpanSys))
	c.setGauge("Mallocs", float64(stats.Mallocs))
	c.setGauge("NextGC", float64(stats.NextGC))
	c.setGauge("NumForcedGC", float64(stats.NumForcedGC))
	c.setGauge("NumGC", float64(stats.NumGC))
	c.setGauge("OtherSys", float64(stats.OtherSys))
	c.setGauge("PauseTotalNs", float64(stats.PauseTotalNs))
	c.setGauge("StackInuse", float64(stats.StackInuse))
	c.setGauge("StackSys", float64(stats.StackSys))
	c.setGauge("Sys", float64(stats.Sys))
	c.setGauge("TotalAlloc", float64(stats.TotalAlloc))
	c.setGauge(randomValueMetric, c.randomFunc())
	c.metrics[pollCountMetric] = Metric{
		Name:         pollCountMetric,
		Type:         models.Counter,
		CounterValue: c.pollCount,
	}
}

// Snapshot возвращает стабильную копию последних метрик.
func (c *Collector) Snapshot() []Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.metrics))
	for name := range c.metrics {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]Metric, 0, len(names))
	for _, name := range names {
		result = append(result, c.metrics[name])
	}

	return result
}

func (c *Collector) setGauge(name string, value float64) {
	c.metrics[name] = Metric{
		Name:       name,
		Type:       models.Gauge,
		GaugeValue: value,
	}
}
