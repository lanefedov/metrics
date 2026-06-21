package agent

import (
	"fmt"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	models "github.com/lanefedov/metrics/internal/model"
)

// GopsutilCollector собирает системные метрики через gopsutil.
type GopsutilCollector struct {
	mu      sync.RWMutex
	metrics map[string]Metric
}

// NewGopsutilCollector создаёт сборщик системных метрик (память и CPU).
func NewGopsutilCollector() *GopsutilCollector {
	return &GopsutilCollector{
		metrics: make(map[string]Metric),
	}
}

// Collect обновляет снимок системных метрик.
func (g *GopsutilCollector) Collect() {
	vmStat, memErr := mem.VirtualMemory()
	cpuPercents, cpuErr := cpu.Percent(0, true)

	g.mu.Lock()
	defer g.mu.Unlock()

	if memErr == nil {
		g.setGauge("TotalMemory", float64(vmStat.Total))
		g.setGauge("FreeMemory", float64(vmStat.Free))
	}

	if cpuErr == nil {
		for i, percent := range cpuPercents {
			g.setGauge(fmt.Sprintf("CPUutilization%d", i+1), percent)
		}
	}
}

// Snapshot возвращает копию последних системных метрик.
func (g *GopsutilCollector) Snapshot() []Metric {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]Metric, 0, len(g.metrics))
	for _, m := range g.metrics {
		result = append(result, m)
	}
	return result
}

func (g *GopsutilCollector) setGauge(name string, value float64) {
	g.metrics[name] = Metric{
		Name:       name,
		Type:       models.Gauge,
		GaugeValue: value,
	}
}
