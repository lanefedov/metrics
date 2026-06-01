package service

import (
	"context"
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
	store       storage.MetricsStorage
	afterUpdate func() error
}

// NewMetricsService создаёт сервис метрик.
func NewMetricsService(store storage.MetricsStorage) *MetricsService {
	return &MetricsService{store: store}
}

// NewMetricsServiceWithAfterUpdate создаёт сервис метрик с колбэком после обновления.
func NewMetricsServiceWithAfterUpdate(store storage.MetricsStorage, afterUpdate func() error) *MetricsService {
	return &MetricsService{
		store:       store,
		afterUpdate: afterUpdate,
	}
}

// UpdateMetric валидирует и сохраняет метрику.
func (s *MetricsService) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	if err := validateMetric(metric); err != nil {
		return err
	}

	if err := s.store.UpdateMetrics(ctx, []models.Metrics{metric}); err != nil {
		return err
	}

	return s.runAfterUpdate()
}

// UpdateMetrics валидирует и сохраняет набор метрик.
func (s *MetricsService) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		if err := validateMetric(metric); err != nil {
			return err
		}
	}

	if err := s.store.UpdateMetrics(ctx, metrics); err != nil {
		return err
	}

	return s.runAfterUpdate()
}

func validateMetric(metric models.Metrics) error {
	if metric.ID == "" {
		return invalidMetric("metric id is required")
	}

	switch metric.MType {
	case models.Counter:
		if metric.Delta == nil {
			return invalidMetric("counter delta is required")
		}
	case models.Gauge:
		if metric.Value == nil {
			return invalidMetric("gauge value is required")
		}
		if math.IsNaN(*metric.Value) || math.IsInf(*metric.Value, 0) {
			return invalidMetric("gauge value must be finite")
		}
	default:
		return invalidMetric("unsupported metric type")
	}

	return nil
}

func (s *MetricsService) runAfterUpdate() error {
	if s.afterUpdate != nil {
		if err := s.afterUpdate(); err != nil {
			return err
		}
	}

	return nil
}

// GetMetric возвращает актуальное значение метрики.
func (s *MetricsService) GetMetric(ctx context.Context, metric models.Metrics) (*models.Metrics, error) {
	if metric.ID == "" {
		return nil, invalidMetric("metric id is required")
	}

	switch metric.MType {
	case models.Counter:
		value, ok, err := s.store.GetCounter(ctx, metric.ID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrMetricNotFound
		}

		return &models.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Delta: &value,
		}, nil
	case models.Gauge:
		value, ok, err := s.store.GetGauge(ctx, metric.ID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrMetricNotFound
		}

		return &models.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: &value,
		}, nil
	default:
		return nil, invalidMetric("unsupported metric type")
	}
}

// ListMetrics возвращает отсортированный срез всех метрик.
func (s *MetricsService) ListMetrics(ctx context.Context) ([]models.Metrics, error) {
	counters, err := s.store.ListCounters(ctx)
	if err != nil {
		return nil, err
	}
	gauges, err := s.store.ListGauges(ctx)
	if err != nil {
		return nil, err
	}

	metrics := make([]models.Metrics, 0, len(counters)+len(gauges))

	for name, value := range counters {
		delta := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &delta,
		})
	}

	for name, value := range gauges {
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

	return metrics, nil
}

func invalidMetric(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidMetric, message)
}
