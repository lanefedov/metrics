package agent

import (
	"context"
	"log"
	"sync"
	"time"
)

type collector interface {
	Collect()
	Snapshot() []Metric
}

type reporter interface {
	Report(metrics []Metric) error
}

// Agent координирует сбор runtime-метрик и их отправку на сервер.
type Agent struct {
	cfg           Config
	collector     collector
	gopsCollector collector
	reporter      reporter
	logger        *log.Logger
}

// New создаёт агент метрик с переданными зависимостями.
func New(cfg Config, collector collector, gopsCollector collector, reporter reporter, logger *log.Logger) *Agent {
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = defaultServerAddress
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaultPollInterval
	}
	if cfg.ReportInterval <= 0 {
		cfg.ReportInterval = defaultReportPeriod
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = defaultRateLimit
	}
	if logger == nil {
		logger = log.Default()
	}

	return &Agent{
		cfg:           cfg,
		collector:     collector,
		gopsCollector: gopsCollector,
		reporter:      reporter,
		logger:        logger,
	}
}

// Run запускает агент: две горутины сбора метрик и worker pool отправки.
// Блокируется до отмены ctx.
func (a *Agent) Run(ctx context.Context) {
	jobs := make(chan []Metric)

	var workerWG sync.WaitGroup
	for range a.cfg.RateLimit {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			for metrics := range jobs {
				if err := a.reporter.Report(metrics); err != nil {
					a.logger.Printf("report metrics: %v", err)
				}
			}
		}()
	}

	go a.runCollector(ctx, a.collector)
	go a.runCollector(ctx, a.gopsCollector)

	reportTicker := time.NewTicker(a.cfg.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			workerWG.Wait()
			return
		case <-reportTicker.C:
			metrics := append(a.collector.Snapshot(), a.gopsCollector.Snapshot()...)
			select {
			case jobs <- metrics:
			case <-ctx.Done():
			}
		}
	}
}

func (a *Agent) runCollector(ctx context.Context, c collector) {
	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.Collect()
		}
	}
}
