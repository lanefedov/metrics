package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoWithSleeperRetriesUntilSuccess(t *testing.T) {
	retriableErr := errors.New("try again")
	var attempts int
	var gotDelays []time.Duration

	err := DoWithSleeper(
		context.Background(),
		DefaultDelays,
		func(err error) bool {
			return errors.Is(err, retriableErr)
		},
		func() error {
			attempts++
			if attempts < 3 {
				return retriableErr
			}
			return nil
		},
		func(_ context.Context, delay time.Duration) error {
			gotDelays = append(gotDelays, delay)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("retry operation: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts: got %d, want 3", attempts)
	}
	assertDelays(t, gotDelays, []time.Duration{time.Second, 3 * time.Second})
}

func TestDoWithSleeperStopsAfterExtraAttempts(t *testing.T) {
	retriableErr := errors.New("try again")
	var attempts int

	err := DoWithSleeper(
		context.Background(),
		DefaultDelays,
		func(err error) bool {
			return errors.Is(err, retriableErr)
		},
		func() error {
			attempts++
			return retriableErr
		},
		func(context.Context, time.Duration) error {
			return nil
		},
	)
	if !errors.Is(err, retriableErr) {
		t.Fatalf("error: got %v, want %v", err, retriableErr)
	}
	if attempts != 4 {
		t.Fatalf("attempts: got %d, want 4", attempts)
	}
}

func TestDoWithSleeperDoesNotRetryNonRetriableError(t *testing.T) {
	nonRetriableErr := errors.New("stop")
	var attempts int

	err := DoWithSleeper(
		context.Background(),
		DefaultDelays,
		func(error) bool {
			return false
		},
		func() error {
			attempts++
			return nonRetriableErr
		},
		func(context.Context, time.Duration) error {
			t.Fatal("sleep should not be called")
			return nil
		},
	)
	if !errors.Is(err, nonRetriableErr) {
		t.Fatalf("error: got %v, want %v", err, nonRetriableErr)
	}
	if attempts != 1 {
		t.Fatalf("attempts: got %d, want 1", attempts)
	}
}

func TestDoWithSleeperStopsWhenContextIsCanceled(t *testing.T) {
	retriableErr := errors.New("try again")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := DoWithSleeper(
		ctx,
		DefaultDelays,
		func(err error) bool {
			return errors.Is(err, retriableErr)
		},
		func() error {
			return retriableErr
		},
		Sleep,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error: got %v, want context.Canceled", err)
	}
}

func assertDelays(t *testing.T, got []time.Duration, want []time.Duration) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("delays: got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("delays: got %v, want %v", got, want)
		}
	}
}
