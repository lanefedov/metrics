package agent

import (
	"log"
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
	cfg       Config
	collector collector
	reporter  reporter
	logger    *log.Logger

	pollElapsed   time.Duration
	reportElapsed time.Duration
}

// New создаёт агент метрик с переданными зависимостями.
func New(cfg Config, collector collector, reporter reporter, logger *log.Logger) *Agent {
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = defaultServerAddress
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaultPollInterval
	}
	if cfg.ReportInterval <= 0 {
		cfg.ReportInterval = defaultReportPeriod
	}
	if cfg.StepInterval <= 0 {
		cfg.StepInterval = defaultStepInterval
	}
	if logger == nil {
		logger = log.Default()
	}

	return &Agent{
		cfg:       cfg,
		collector: collector,
		reporter:  reporter,
		logger:    logger,
	}
}

// Run запускает бесконечный цикл работы агента.
func (a *Agent) Run() {
	for {
		a.Step()
		time.Sleep(a.cfg.StepInterval)
	}
}

// Step продвигает состояние агента на один шаг.
func (a *Agent) Step() {
	a.pollElapsed += a.cfg.StepInterval
	a.reportElapsed += a.cfg.StepInterval

	if a.pollElapsed >= a.cfg.PollInterval {
		a.collector.Collect()
		a.pollElapsed = 0
	}

	if a.reportElapsed >= a.cfg.ReportInterval {
		if err := a.reporter.Report(a.collector.Snapshot()); err != nil {
			a.logger.Printf("report metrics: %v", err)
		}
		a.reportElapsed = 0
	}
}
