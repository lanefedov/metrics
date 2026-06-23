package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	models "github.com/lanefedov/metrics/internal/model"
	"github.com/lanefedov/metrics/internal/retry"
)

// PostgresStorage хранит метрики в PostgreSQL.
type PostgresStorage struct {
	pool *pgxpool.Pool
}

// NewPostgresStorage создаёт PostgreSQL-хранилище поверх готового пула соединений.
func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{pool: pool}
}

// SetGauge устанавливает значение метрики типа gauge.
func (s *PostgresStorage) SetGauge(ctx context.Context, name string, value float64) error {
	return s.retryPostgres(ctx, func() error {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO metrics (metric_type, name, gauge_value)
			VALUES ('gauge', $1, $2)
			ON CONFLICT (metric_type, name)
			DO UPDATE SET gauge_value = EXCLUDED.gauge_value
		`, name, value)
		if err != nil {
			return fmt.Errorf("set gauge: %w", err)
		}

		return nil
	})
}

// AddCounter добавляет delta к накопленному значению counter.
func (s *PostgresStorage) AddCounter(ctx context.Context, name string, delta int64) error {
	return s.retryPostgres(ctx, func() error {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO metrics (metric_type, name, counter_value)
			VALUES ('counter', $1, $2)
			ON CONFLICT (metric_type, name)
			DO UPDATE SET counter_value = metrics.counter_value + EXCLUDED.counter_value
		`, name, delta)
		if err != nil {
			return fmt.Errorf("add counter: %w", err)
		}

		return nil
	})
}

// UpdateMetrics применяет набор обновлений метрик в одной транзакции.
func (s *PostgresStorage) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	return s.retryPostgres(ctx, func() error {
		return s.updateMetricsOnce(ctx, metrics)
	})
}

func (s *PostgresStorage) updateMetricsOnce(ctx context.Context, metrics []models.Metrics) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update metrics transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			_, err = tx.Exec(ctx, `
				INSERT INTO metrics (metric_type, name, counter_value)
				VALUES ('counter', $1, $2)
				ON CONFLICT (metric_type, name)
				DO UPDATE SET counter_value = metrics.counter_value + EXCLUDED.counter_value
			`, metric.ID, *metric.Delta)
			if err != nil {
				return fmt.Errorf("add counter in batch: %w", err)
			}
		case models.Gauge:
			_, err = tx.Exec(ctx, `
				INSERT INTO metrics (metric_type, name, gauge_value)
				VALUES ('gauge', $1, $2)
				ON CONFLICT (metric_type, name)
				DO UPDATE SET gauge_value = EXCLUDED.gauge_value
			`, metric.ID, *metric.Value)
			if err != nil {
				return fmt.Errorf("set gauge in batch: %w", err)
			}
		default:
			return fmt.Errorf("unsupported metric type %q", metric.MType)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update metrics transaction: %w", err)
	}

	return nil
}

// GetGauge возвращает текущее значение gauge и признак его существования.
func (s *PostgresStorage) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	var value float64
	var ok bool
	err := s.retryPostgres(ctx, func() error {
		err := s.pool.QueryRow(ctx, `
			SELECT gauge_value
			FROM metrics
			WHERE metric_type = 'gauge' AND name = $1
		`, name).Scan(&value)
		if errors.Is(err, pgx.ErrNoRows) {
			ok = false
			return nil
		}
		if err != nil {
			return fmt.Errorf("get gauge: %w", err)
		}

		ok = true
		return nil
	})
	if err != nil {
		return 0, false, err
	}

	return value, ok, nil
}

// GetCounter возвращает текущее значение counter и признак его существования.
func (s *PostgresStorage) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	var value int64
	var ok bool
	err := s.retryPostgres(ctx, func() error {
		err := s.pool.QueryRow(ctx, `
			SELECT counter_value
			FROM metrics
			WHERE metric_type = 'counter' AND name = $1
		`, name).Scan(&value)
		if errors.Is(err, pgx.ErrNoRows) {
			ok = false
			return nil
		}
		if err != nil {
			return fmt.Errorf("get counter: %w", err)
		}

		ok = true
		return nil
	})
	if err != nil {
		return 0, false, err
	}

	return value, ok, nil
}

// ListGauges возвращает снимок всех gauge-метрик.
func (s *PostgresStorage) ListGauges(ctx context.Context) (map[string]float64, error) {
	var gauges map[string]float64
	err := s.retryPostgres(ctx, func() error {
		rows, err := s.pool.Query(ctx, `
			SELECT name, gauge_value
			FROM metrics
			WHERE metric_type = 'gauge'
		`)
		if err != nil {
			return fmt.Errorf("list gauges: %w", err)
		}
		defer rows.Close()

		result := make(map[string]float64)
		for rows.Next() {
			var name string
			var value float64
			if err := rows.Scan(&name, &value); err != nil {
				return fmt.Errorf("scan gauge: %w", err)
			}
			result[name] = value
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate gauges: %w", err)
		}

		gauges = result
		return nil
	})
	if err != nil {
		return nil, err
	}

	return gauges, nil
}

// ListCounters возвращает снимок всех counter-метрик.
func (s *PostgresStorage) ListCounters(ctx context.Context) (map[string]int64, error) {
	var counters map[string]int64
	err := s.retryPostgres(ctx, func() error {
		rows, err := s.pool.Query(ctx, `
			SELECT name, counter_value
			FROM metrics
			WHERE metric_type = 'counter'
		`)
		if err != nil {
			return fmt.Errorf("list counters: %w", err)
		}
		defer rows.Close()

		result := make(map[string]int64)
		for rows.Next() {
			var name string
			var value int64
			if err := rows.Scan(&name, &value); err != nil {
				return fmt.Errorf("scan counter: %w", err)
			}
			result[name] = value
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate counters: %w", err)
		}

		counters = result
		return nil
	})
	if err != nil {
		return nil, err
	}

	return counters, nil
}

func (s *PostgresStorage) retryPostgres(ctx context.Context, operation func() error) error {
	return retry.Do(ctx, isRetriablePostgresError, operation)
}

func isRetriablePostgresError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return isPostgresConnectionExceptionCode(pgErr.Code)
}

func isPostgresConnectionExceptionCode(code string) bool {
	return strings.HasPrefix(code, "08")
}
