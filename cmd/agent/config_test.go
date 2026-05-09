package main

import (
	"testing"
	"time"
)

func TestLoadAgentConfigUsesDefaultsWithoutEnv(t *testing.T) {
	cfg, err := loadAgentConfigWithEnv(nil, emptyEnvLookup)
	if err != nil {
		t.Fatalf("load agent config: %v", err)
	}

	if cfg.ServerAddress != "localhost:8080" {
		t.Fatalf("address: got %q, want %q", cfg.ServerAddress, "localhost:8080")
	}
	if cfg.ReportInterval != 10*time.Second {
		t.Fatalf("report interval: got %s, want %s", cfg.ReportInterval, 10*time.Second)
	}
	if cfg.PollInterval != 2*time.Second {
		t.Fatalf("poll interval: got %s, want %s", cfg.PollInterval, 2*time.Second)
	}
}

func TestLoadAgentConfigUsesEnvValues(t *testing.T) {
	cfg, err := loadAgentConfigWithEnv(nil, envLookup(map[string]string{
		addressEnvKey:        "localhost:9090",
		reportIntervalEnvKey: "13",
		pollIntervalEnvKey:   "5",
	}))
	if err != nil {
		t.Fatalf("load agent config: %v", err)
	}

	if cfg.ServerAddress != "localhost:9090" {
		t.Fatalf("address: got %q, want %q", cfg.ServerAddress, "localhost:9090")
	}
	if cfg.ReportInterval != 13*time.Second {
		t.Fatalf("report interval: got %s, want %s", cfg.ReportInterval, 13*time.Second)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Fatalf("poll interval: got %s, want %s", cfg.PollInterval, 5*time.Second)
	}
}

func TestLoadAgentConfigEnvOverridesFlags(t *testing.T) {
	cfg, err := loadAgentConfigWithEnv([]string{
		"-a=127.0.0.1:9000",
		"-r=15",
		"-p=4",
	}, envLookup(map[string]string{
		addressEnvKey:        "localhost:9090",
		reportIntervalEnvKey: "13",
		pollIntervalEnvKey:   "5",
	}))
	if err != nil {
		t.Fatalf("load agent config: %v", err)
	}

	if cfg.ServerAddress != "localhost:9090" {
		t.Fatalf("address: got %q, want %q", cfg.ServerAddress, "localhost:9090")
	}
	if cfg.ReportInterval != 13*time.Second {
		t.Fatalf("report interval: got %s, want %s", cfg.ReportInterval, 13*time.Second)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Fatalf("poll interval: got %s, want %s", cfg.PollInterval, 5*time.Second)
	}
}

func TestLoadAgentConfigRejectsInvalidEnvInterval(t *testing.T) {
	_, err := loadAgentConfigWithEnv(nil, envLookup(map[string]string{
		reportIntervalEnvKey: "ten",
	}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "REPORT_INTERVAL must be an integer number of seconds" {
		t.Fatalf("error: got %q, want %q", err.Error(), "REPORT_INTERVAL must be an integer number of seconds")
	}
}

func emptyEnvLookup(string) (string, bool) {
	return "", false
}

func envLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
