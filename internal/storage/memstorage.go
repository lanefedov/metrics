package storage

import "sync"

// MetricsUpdater описывает операции обновления метрик.
type MetricsUpdater interface {
	SetGauge(name string, value float64)
	AddCounter(name string, delta int64)
}

// MetricsReader описывает операции чтения отдельных метрик.
type MetricsReader interface {
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}

// MetricsLister описывает операции получения снимка всех метрик.
type MetricsLister interface {
	ListGauges() map[string]float64
	ListCounters() map[string]int64
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
func (m *MemStorage) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[name] = value
}

// AddCounter добавляет delta к накопленному значению counter.
func (m *MemStorage) AddCounter(name string, delta int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[name] += delta
}

// GetGauge возвращает текущее значение gauge и признак его существования.
func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.gauges[name]
	return value, ok
}

// GetCounter возвращает текущее значение counter и признак его существования.
func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.counters[name]
	return value, ok
}

// ListGauges возвращает снимок всех gauge-метрик.
func (m *MemStorage) ListGauges() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gauges := make(map[string]float64, len(m.gauges))
	for name, value := range m.gauges {
		gauges[name] = value
	}

	return gauges
}

// ListCounters возвращает снимок всех counter-метрик.
func (m *MemStorage) ListCounters() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counters := make(map[string]int64, len(m.counters))
	for name, value := range m.counters {
		counters[name] = value
	}

	return counters
}

func (m *MemStorage) replaceAll(counters map[string]int64, gauges map[string]float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters = counters
	m.gauges = gauges
}
