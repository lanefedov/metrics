package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadAgentConfigUsesDefaultsWithoutEnv(t *testing.T) {
	unsetAgentEnv(t)

	cfg, err := loadAgentConfig(nil)
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
	unsetAgentEnv(t)
	t.Setenv("ADDRESS", "localhost:9090")
	t.Setenv("REPORT_INTERVAL", "13")
	t.Setenv("POLL_INTERVAL", "5")

	cfg, err := loadAgentConfig(nil)
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
	unsetAgentEnv(t)
	t.Setenv("ADDRESS", "localhost:9090")
	t.Setenv("REPORT_INTERVAL", "13")
	t.Setenv("POLL_INTERVAL", "5")

	cfg, err := loadAgentConfig([]string{
		"-a=127.0.0.1:9000",
		"-r=15",
		"-p=4",
	})
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
	unsetAgentEnv(t)
	t.Setenv("REPORT_INTERVAL", "ten")

	_, err := loadAgentConfig(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "REPORT_INTERVAL") {
		t.Fatalf("error: got %q, want it to mention REPORT_INTERVAL", err.Error())
	}
}

func unsetAgentEnv(t *testing.T) {
	t.Helper()

	keys := []string{"ADDRESS", "REPORT_INTERVAL", "POLL_INTERVAL"}
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}

		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(key, value)
				return
			}
			_ = os.Unsetenv(key)
		})
	}
}
