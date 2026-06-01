package storage

import (
	"context"
	"fmt"
	"sync"

	models "github.com/lanefedov/metrics/internal/model"
)

// MetricsUpdater описывает операции обновления метрик.
type MetricsUpdater interface {
	SetGauge(ctx context.Context, name string, value float64) error
	AddCounter(ctx context.Context, name string, delta int64) error
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) error
}

// MetricsReader описывает операции чтения отдельных метрик.
type MetricsReader interface {
	GetGauge(ctx context.Context, name string) (float64, bool, error)
	GetCounter(ctx context.Context, name string) (int64, bool, error)
}

// MetricsLister описывает операции получения снимка всех метрик.
type MetricsLister interface {
	ListGauges(ctx context.Context) (map[string]float64, error)
	ListCounters(ctx context.Context) (map[string]int64, error)
}

// MetricsStorage объединяет операции чтения и записи метрик.
type MetricsStorage interface {
	MetricsUpdater
	MetricsReader
	MetricsLister
}

// MemStorage хранит метрики типов gauge и counter в памяти.
type MemStorage struct {
	mu       sync.RWMutex
	counters map[string]int64
	gauges   map[string]float64
}

// NewMemStorage создаёт пустое хранилище в памяти.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

// SetGauge устанавливает значение метрики типа gauge.
func (m *MemStorage) SetGauge(_ context.Context, name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[name] = value
	return nil
}

// AddCounter добавляет delta к накопленному значению counter.
func (m *MemStorage) AddCounter(_ context.Context, name string, delta int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[name] += delta
	return nil
}

// UpdateMetrics применяет набор обновлений метрик под одной блокировкой.
func (m *MemStorage) UpdateMetrics(_ context.Context, metrics []models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			m.counters[metric.ID] += *metric.Delta
		case models.Gauge:
			m.gauges[metric.ID] = *metric.Value
		default:
			return fmt.Errorf("unsupported metric type %q", metric.MType)
		}
	}

	return nil
}

// GetGauge возвращает текущее значение gauge и признак его существования.
func (m *MemStorage) GetGauge(_ context.Context, name string) (float64, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.gauges[name]
	return value, ok, nil
}

// GetCounter возвращает текущее значение counter и признак его существования.
func (m *MemStorage) GetCounter(_ context.Context, name string) (int64, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.counters[name]
	return value, ok, nil
}

// ListGauges возвращает снимок всех gauge-метрик.
func (m *MemStorage) ListGauges(_ context.Context) (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gauges := make(map[string]float64, len(m.gauges))
	for name, value := range m.gauges {
		gauges[name] = value
	}

	return gauges, nil
}

// ListCounters возвращает снимок всех counter-метрик.
func (m *MemStorage) ListCounters(_ context.Context) (map[string]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counters := make(map[string]int64, len(m.counters))
	for name, value := range m.counters {
		counters[name] = value
	}

	return counters, nil
}

func (m *MemStorage) replaceAll(counters map[string]int64, gauges map[string]float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters = counters
	m.gauges = gauges
}
