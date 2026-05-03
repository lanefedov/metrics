package storage

// MetricsStorage описывает операции обновления метрик в памяти.
type MetricsStorage interface {
	SetGauge(name string, value float64)
	AddCounter(name string, delta int64)
}

// MemStorage хранит gauge и counter в памяти.
type MemStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

// NewMemStorage создаёт пустое in-memory-хранилище.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

// SetGauge устанавливает значение gauge.
func (m *MemStorage) SetGauge(name string, value float64) {
	m.gauges[name] = value
}

// AddCounter добавляет delta к накопленному counter.
func (m *MemStorage) AddCounter(name string, delta int64) {
	m.counters[name] += delta
}
