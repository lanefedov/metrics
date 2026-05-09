package main

import "testing"

func TestLoadServerConfigUsesDefaultsWithoutEnv(t *testing.T) {
	cfg, err := loadServerConfigWithEnv(nil, emptyEnvLookup)
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}

	if cfg.address != defaultListenAddress {
		t.Fatalf("address: got %q, want %q", cfg.address, defaultListenAddress)
	}
}

func TestLoadServerConfigUsesEnvValue(t *testing.T) {
	cfg, err := loadServerConfigWithEnv(nil, envLookup(map[string]string{
		addressEnvKey: "localhost:9090",
	}))
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}

	if cfg.address != "localhost:9090" {
		t.Fatalf("address: got %q, want %q", cfg.address, "localhost:9090")
	}
}

func TestLoadServerConfigEnvOverridesFlag(t *testing.T) {
	cfg, err := loadServerConfigWithEnv([]string{"-a=127.0.0.1:9000"}, envLookup(map[string]string{
		addressEnvKey: "localhost:9090",
	}))
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}

	if cfg.address != "localhost:9090" {
		t.Fatalf("address: got %q, want %q", cfg.address, "localhost:9090")
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
