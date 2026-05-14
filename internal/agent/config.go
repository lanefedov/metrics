package agent

import "time"

const (
	defaultServerAddress = "localhost:8080"
	defaultPollInterval  = 2 * time.Second
	defaultReportPeriod  = 10 * time.Second
	defaultStepInterval  = time.Second
)

// Config хранит параметры запуска агента метрик.
type Config struct {
	ServerAddress  string
	PollInterval   time.Duration
	ReportInterval time.Duration
	StepInterval   time.Duration
}

// DefaultConfig возвращает значения по умолчанию для HTTP-агента.
func DefaultConfig() Config {
	return Config{
		ServerAddress:  defaultServerAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportPeriod,
		StepInterval:   defaultStepInterval,
	}
}
