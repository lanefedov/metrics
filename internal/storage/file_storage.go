package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	models "github.com/lanefedov/metrics/internal/model"
)

// SaveToFile сериализует все текущие метрики в JSON-файл.
func (m *MemStorage) SaveToFile(path string) error {
	metrics := m.snapshotMetrics()

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create storage directory: %w", err)
		}
	}

	tempFile, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary storage file: %w", err)
	}

	tempPath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

	encoder := json.NewEncoder(tempFile)
	if err := encoder.Encode(metrics); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("encode metrics: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temporary storage file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace storage file: %w", err)
	}

	return nil
}

// LoadFromFile восстанавливает метрики из ранее сохранённого JSON-файла.
func (m *MemStorage) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read storage file: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		m.replaceAll(make(map[string]int64), make(map[string]float64))
		return nil
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("decode metrics: %w", err)
	}

	counters := make(map[string]int64)
	gauges := make(map[string]float64)

	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("counter %q has no delta", metric.ID)
			}
			counters[metric.ID] = *metric.Delta
		case models.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("gauge %q has no value", metric.ID)
			}
			gauges[metric.ID] = *metric.Value
		default:
			return fmt.Errorf("unsupported metric type %q", metric.MType)
		}
	}

	m.replaceAll(counters, gauges)
	return nil
}

func (m *MemStorage) snapshotMetrics() []models.Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(m.counters)+len(m.gauges))
	for name, value := range m.counters {
		delta := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &delta,
		})
	}

	for name, value := range m.gauges {
		gaugeValue := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &gaugeValue,
		})
	}

	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].MType == metrics[j].MType {
			return metrics[i].ID < metrics[j].ID
		}
		return metrics[i].MType < metrics[j].MType
	})

	return metrics
}
