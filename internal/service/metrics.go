package service

import (
	"errors"
	"fmt"
	"math"
	"sort"

	models "github.com/lanefedov/metrics/internal/model"
	"github.com/lanefedov/metrics/internal/storage"
)

var (
	// ErrInvalidMetric сигнализирует о невалидных данных метрики.
	ErrInvalidMetric = errors.New("invalid metric")
	// ErrMetricNotFound сигнализирует об отсутствии метрики в хранилище.
	ErrMetricNotFound = errors.New("metric not found")
)

// MetricsService содержит бизнес-логику работы с метриками.
type MetricsService struct {
	store storage.MetricsStorage
}

// NewMetricsService создаёт сервис метрик.
func NewMetricsService(store storage.MetricsStorage) *MetricsService {
	return &MetricsService{store: store}
}

// UpdateMetric валидирует и сохраняет метрику.
func (s *MetricsService) UpdateMetric(metric models.Metrics) error {
	if metric.ID == "" {
		return invalidMetric("metric id is required")
	}

	switch metric.MType {
	case models.Counter:
		if metric.Delta == nil {
			return invalidMetric("counter delta is required")
		}
		s.store.AddCounter(metric.ID, *metric.Delta)
		return nil
	case models.Gauge:
		if metric.Value == nil {
			return invalidMetric("gauge value is required")
		}
		if math.IsNaN(*metric.Value) || math.IsInf(*metric.Value, 0) {
			return invalidMetric("gauge value must be finite")
		}
		s.store.SetGauge(metric.ID, *metric.Value)
		return nil
	default:
		return invalidMetric("unsupported metric type")
	}
}

// GetMetric возвращает актуальное значение метрики.
func (s *MetricsService) GetMetric(metric models.Metrics) (models.Metrics, error) {
	if metric.ID == "" {
		return models.Metrics{}, invalidMetric("metric id is required")
	}

	switch metric.MType {
	case models.Counter:
		value, ok := s.store.GetCounter(metric.ID)
		if !ok {
			return models.Metrics{}, ErrMetricNotFound
		}

		return models.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Delta: &value,
		}, nil
	case models.Gauge:
		value, ok := s.store.GetGauge(metric.ID)
		if !ok {
			return models.Metrics{}, ErrMetricNotFound
		}

		return models.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: &value,
		}, nil
	default:
		return models.Metrics{}, invalidMetric("unsupported metric type")
	}
}

// ListMetrics возвращает отсортированный срез всех метрик.
func (s *MetricsService) ListMetrics() []models.Metrics {
	metrics := make([]models.Metrics, 0, len(s.store.ListCounters())+len(s.store.ListGauges()))

	for name, value := range s.store.ListCounters() {
		delta := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &delta,
		})
	}

	for name, value := range s.store.ListGauges() {
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

func invalidMetric(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidMetric, message)
}
