package agent

import (
	"io"
	"log"
	"testing"
	"time"

	models "github.com/lanefedov/metrics/internal/model"
)

func TestAgentStepTriggersPollAndReportByIntervals(t *testing.T) {
	collector := &fakeCollector{
		snapshot: []Metric{
			{Name: "PollCount", Type: models.Counter, CounterValue: 3},
		},
	}
	reporter := &fakeReporter{}
	app := New(
		Config{
			ServerAddress:  "http://localhost:8080",
			PollInterval:   2 * time.Second,
			ReportInterval: 5 * time.Second,
			StepInterval:   time.Second,
		},
		collector,
		reporter,
		log.New(io.Discard, "", 0),
	)

	for range 4 {
		app.Step()
	}

	if collector.collectCalls != 2 {
		t.Fatalf("collect calls after 4 steps: got %d, want 2", collector.collectCalls)
	}
	if reporter.reportCalls != 0 {
		t.Fatalf("report calls after 4 steps: got %d, want 0", reporter.reportCalls)
	}

	app.Step()

	if collector.collectCalls != 2 {
		t.Fatalf("collect calls after 5 steps: got %d, want 2", collector.collectCalls)
	}
	if reporter.reportCalls != 1 {
		t.Fatalf("report calls after 5 steps: got %d, want 1", reporter.reportCalls)
	}
	if len(reporter.lastMetrics) != 1 || reporter.lastMetrics[0].Name != "PollCount" {
		t.Fatalf("reported metrics: got %+v", reporter.lastMetrics)
	}
	if collector.snapshotCalls != 1 {
		t.Fatalf("snapshot calls: got %d, want 1", collector.snapshotCalls)
	}
}

func TestAgentStepLogsReportErrorAndContinues(t *testing.T) {
	collector := &fakeCollector{}
	reporter := &fakeReporter{err: assertiveError("send failed")}
	app := New(
		Config{
			PollInterval:   time.Second,
			ReportInterval: time.Second,
			StepInterval:   time.Second,
		},
		collector,
		reporter,
		log.New(io.Discard, "", 0),
	)

	app.Step()

	if collector.collectCalls != 1 {
		t.Fatalf("collect calls: got %d, want 1", collector.collectCalls)
	}
	if reporter.reportCalls != 1 {
		t.Fatalf("report calls: got %d, want 1", reporter.reportCalls)
	}
}

type fakeCollector struct {
	collectCalls  int
	snapshotCalls int
	snapshot      []Metric
}

func (f *fakeCollector) Collect() {
	f.collectCalls++
}

func (f *fakeCollector) Snapshot() []Metric {
	f.snapshotCalls++
	return append([]Metric(nil), f.snapshot...)
}

type fakeReporter struct {
	reportCalls int
	lastMetrics []Metric
	err         error
}

func (f *fakeReporter) Report(metrics []Metric) error {
	f.reportCalls++
	f.lastMetrics = append([]Metric(nil), metrics...)
	return f.err
}
