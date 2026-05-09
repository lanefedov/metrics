package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/lanefedov/metrics/internal/agent"
)

const (
	addressEnvKey        = "ADDRESS"
	reportIntervalEnvKey = "REPORT_INTERVAL"
	pollIntervalEnvKey   = "POLL_INTERVAL"
)

func loadAgentConfig(args []string) (agent.Config, error) {
	return loadAgentConfigWithEnv(args, os.LookupEnv)
}

func loadAgentConfigWithEnv(args []string, lookupEnv func(string) (string, bool)) (agent.Config, error) {
	cfg, err := parseAgentFlags(args)
	if err != nil {
		return agent.Config{}, err
	}

	if value, ok := lookupEnv(addressEnvKey); ok {
		cfg.ServerAddress = value
	}

	if value, ok := lookupEnv(reportIntervalEnvKey); ok {
		seconds, err := parseEnvSeconds(reportIntervalEnvKey, value)
		if err != nil {
			return agent.Config{}, err
		}
		cfg.ReportInterval = time.Duration(seconds) * time.Second
	}

	if value, ok := lookupEnv(pollIntervalEnvKey); ok {
		seconds, err := parseEnvSeconds(pollIntervalEnvKey, value)
		if err != nil {
			return agent.Config{}, err
		}
		cfg.PollInterval = time.Duration(seconds) * time.Second
	}

	return cfg, nil
}

func parseEnvSeconds(key string, value string) (int, error) {
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer number of seconds", key)
	}

	return seconds, nil
}
