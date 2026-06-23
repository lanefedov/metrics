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
	metrics []Metric
}

// NewGopsutilCollector создаёт сборщик системных метрик (память и CPU).
func NewGopsutilCollector() *GopsutilCollector {
	return &GopsutilCollector{}
}

// Collect обновляет снимок системных метрик.
func (g *GopsutilCollector) Collect() {
	vmStat, memErr := mem.VirtualMemory()
	cpuPercents, cpuErr := cpu.Percent(0, true)

	g.mu.Lock()
	defer g.mu.Unlock()

	g.metrics = g.metrics[:0]

	if memErr == nil {
		g.metrics = append(g.metrics,
			Metric{Name: "TotalMemory", Type: models.Gauge, GaugeValue: float64(vmStat.Total)},
			Metric{Name: "FreeMemory", Type: models.Gauge, GaugeValue: float64(vmStat.Free)},
		)
	}

	if cpuErr == nil {
		for i, percent := range cpuPercents {
			g.metrics = append(g.metrics, Metric{
				Name:       fmt.Sprintf("CPUutilization%d", i+1),
				Type:       models.Gauge,
				GaugeValue: percent,
			})
		}
	}
}

// Snapshot возвращает копию последних системных метрик.
func (g *GopsutilCollector) Snapshot() []Metric {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]Metric, len(g.metrics))
	copy(result, g.metrics)
	return result
}
