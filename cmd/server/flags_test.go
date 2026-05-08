package main

import "testing"

func TestParseServerFlagsUsesDefaults(t *testing.T) {
	cfg, err := parseServerFlags(nil)
	if err != nil {
		t.Fatalf("parse server flags: %v", err)
	}

	if cfg.address != defaultListenAddress {
		t.Fatalf("address: got %q, want %q", cfg.address, defaultListenAddress)
	}
}

func TestParseServerFlagsOverridesAddress(t *testing.T) {
	cfg, err := parseServerFlags([]string{"-a=127.0.0.1:9000"})
	if err != nil {
		t.Fatalf("parse server flags: %v", err)
	}

	if cfg.address != "127.0.0.1:9000" {
		t.Fatalf("address: got %q, want %q", cfg.address, "127.0.0.1:9000")
	}
}

func TestParseServerFlagsRejectsUnknownFlags(t *testing.T) {
	_, err := parseServerFlags([]string{"-x=1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "flag provided but not defined: -x" {
		t.Fatalf("error: got %q, want %q", err.Error(), "flag provided but not defined: -x")
	}
}
