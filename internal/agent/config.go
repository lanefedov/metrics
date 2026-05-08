package agent

import "time"

const (
	defaultServerAddress = "http://localhost:8080"
	defaultPollInterval  = 2 * time.Second
	defaultReportPeriod  = 10 * time.Second
	defaultStepInterval  = time.Second
)

// Config stores the runtime settings of the metrics agent.
type Config struct {
	ServerAddress  string
	PollInterval   time.Duration
	ReportInterval time.Duration
	StepInterval   time.Duration
}

// DefaultConfig returns the task defaults for the HTTP agent.
func DefaultConfig() Config {
	return Config{
		ServerAddress:  defaultServerAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportPeriod,
		StepInterval:   defaultStepInterval,
	}
}
