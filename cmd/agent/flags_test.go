package main

import (
	"testing"
	"time"
)

func TestParseAgentFlagsUsesDefaults(t *testing.T) {
	cfg, err := parseAgentFlags(nil)
	if err != nil {
		t.Fatalf("parse agent flags: %v", err)
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

func TestParseAgentFlagsOverridesValues(t *testing.T) {
	cfg, err := parseAgentFlags([]string{
		"-a=127.0.0.1:9000",
		"-r=15",
		"-p=4",
	})
	if err != nil {
		t.Fatalf("parse agent flags: %v", err)
	}

	if cfg.ServerAddress != "127.0.0.1:9000" {
		t.Fatalf("address: got %q, want %q", cfg.ServerAddress, "127.0.0.1:9000")
	}
	if cfg.ReportInterval != 15*time.Second {
		t.Fatalf("report interval: got %s, want %s", cfg.ReportInterval, 15*time.Second)
	}
	if cfg.PollInterval != 4*time.Second {
		t.Fatalf("poll interval: got %s, want %s", cfg.PollInterval, 4*time.Second)
	}
}

func TestParseAgentFlagsRejectsUnknownFlags(t *testing.T) {
	_, err := parseAgentFlags([]string{"-x=1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "flag provided but not defined: -x" {
		t.Fatalf("error: got %q, want %q", err.Error(), "flag provided but not defined: -x")
	}
}
