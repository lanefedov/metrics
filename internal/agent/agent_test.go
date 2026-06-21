package agent

import (
	"context"
	"io"
	"log"
	"sync"
	"testing"
	"time"

	models "github.com/lanefedov/metrics/internal/model"
)

func TestAgentRunCollectsAndReports(t *testing.T) {
	collector := &fakeCollector{
		snapshot: []Metric{
			{Name: "PollCount", Type: models.Counter, CounterValue: 3},
		},
	}
	gopsCollector := &fakeCollector{}
	reporter := &fakeReporter{}

	app := New(
		Config{
			PollInterval:   20 * time.Millisecond,
			ReportInterval: 60 * time.Millisecond,
			RateLimit:      2,
		},
		collector,
		gopsCollector,
		reporter,
		log.New(io.Discard, "", 0),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	app.Run(ctx)

	// За 200ms при poll=20ms ожидается минимум 8 вызовов Collect
	if collector.getCollectCalls() < 5 {
		t.Fatalf("runtime collect calls: got %d, want at least 5", collector.getCollectCalls())
	}
	if gopsCollector.getCollectCalls() < 5 {
		t.Fatalf("gops collect calls: got %d, want at least 5", gopsCollector.getCollectCalls())
	}
	// За 200ms при report=60ms ожидается минимум 2 отчёта
	if reporter.getReportCalls() < 2 {
		t.Fatalf("report calls: got %d, want at least 2", reporter.getReportCalls())
	}
	// Проверяем, что метрики runtime попали в отчёт
	last := reporter.getLastMetrics()
	found := false
	for _, m := range last {
		if m.Name == "PollCount" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("PollCount not found in last report: %+v", last)
	}
}

func TestAgentRunLogsReportErrorAndContinues(t *testing.T) {
	collector := &fakeCollector{}
	gopsCollector := &fakeCollector{}
	reporter := &fakeReporter{err: assertiveError("send failed")}

	app := New(
		Config{
			PollInterval:   20 * time.Millisecond,
			ReportInterval: 30 * time.Millisecond,
			RateLimit:      1,
		},
		collector,
		gopsCollector,
		reporter,
		log.New(io.Discard, "", 0),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	app.Run(ctx)

	// Агент должен продолжать вызывать Report несмотря на ошибки
	if reporter.getReportCalls() < 2 {
		t.Fatalf("report calls: got %d, want at least 2", reporter.getReportCalls())
	}
}

// fakeCollector — потокобезопасный тестовый сборщик метрик.
type fakeCollector struct {
	mu            sync.Mutex
	collectCalls  int
	snapshotCalls int
	snapshot      []Metric
}

func (f *fakeCollector) Collect() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.collectCalls++
}

func (f *fakeCollector) Snapshot() []Metric {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.snapshotCalls++
	return append([]Metric(nil), f.snapshot...)
}

func (f *fakeCollector) getCollectCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.collectCalls
}

// fakeReporter — потокобезопасный тестовый отправитель метрик.
type fakeReporter struct {
	mu          sync.Mutex
	reportCalls int
	lastMetrics []Metric
	err         error
}

func (f *fakeReporter) Report(metrics []Metric) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reportCalls++
	f.lastMetrics = append([]Metric(nil), metrics...)
	return f.err
}

func (f *fakeReporter) getReportCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reportCalls
}

func (f *fakeReporter) getLastMetrics() []Metric {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Metric(nil), f.lastMetrics...)
}
