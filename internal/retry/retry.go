package retry

import (
	"context"
	"time"
)

// DefaultDelays задаёт три дополнительные попытки с растущими интервалами.
var DefaultDelays = []time.Duration{
	time.Second,
	3 * time.Second,
	5 * time.Second,
}

// ShouldRetry определяет, стоит ли повторять операцию после ошибки.
type ShouldRetry func(error) bool

// SleepFunc ждёт указанный интервал или завершает ожидание по контексту.
type SleepFunc func(context.Context, time.Duration) error

// Do выполняет операцию и повторяет её для retriable-ошибок.
func Do(ctx context.Context, shouldRetry ShouldRetry, operation func() error) error {
	return DoWithSleeper(ctx, DefaultDelays, shouldRetry, operation, Sleep)
}

// DoWithSleeper выполняет операцию с настраиваемыми задержками.
func DoWithSleeper(
	ctx context.Context,
	delays []time.Duration,
	shouldRetry ShouldRetry,
	operation func() error,
	sleep SleepFunc,
) error {
	if sleep == nil {
		sleep = Sleep
	}

	for attempt := 0; ; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		if shouldRetry == nil || !shouldRetry(err) || attempt >= len(delays) {
			return err
		}
		if sleepErr := sleep(ctx, delays[attempt]); sleepErr != nil {
			return sleepErr
		}
	}
}

// Sleep ждёт delay или возвращает ошибку отменённого контекста.
func Sleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
