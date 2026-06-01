package main

import (
	"testing"
	"time"
)

func TestLoadServerConfigUsesDefaultsWithoutEnv(t *testing.T) {
	cfg, err := loadServerConfigWithEnv(nil, emptyEnvLookup)
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}

	if cfg.address != defaultListenAddress {
		t.Fatalf("address: got %q, want %q", cfg.address, defaultListenAddress)
	}
	if cfg.storeInterval != defaultStoreInterval {
		t.Fatalf("store interval: got %v, want %v", cfg.storeInterval, defaultStoreInterval)
	}
	if cfg.fileStoragePath != defaultStoreFilePath {
		t.Fatalf("file storage path: got %q, want %q", cfg.fileStoragePath, defaultStoreFilePath)
	}
	if !cfg.restore {
		t.Fatal("restore: got false, want true")
	}
	if cfg.databaseDSN != "" {
		t.Fatalf("database dsn: got %q, want empty", cfg.databaseDSN)
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

func TestLoadServerConfigUsesEnvValues(t *testing.T) {
	const dsn = "host=localhost port=5432 user=test dbname=metrics sslmode=disable"

	cfg, err := loadServerConfigWithEnv(nil, envLookup(map[string]string{
		addressEnvKey:         "localhost:9090",
		storeIntervalEnvKey:   "1",
		fileStoragePathEnvKey: "/tmp/server-db.json",
		restoreEnvKey:         "true",
		databaseDSNEnvKey:     dsn,
	}))
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}

	if cfg.address != "localhost:9090" {
		t.Fatalf("address: got %q, want %q", cfg.address, "localhost:9090")
	}
	if cfg.storeInterval != time.Second {
		t.Fatalf("store interval: got %v, want %v", cfg.storeInterval, time.Second)
	}
	if cfg.fileStoragePath != "/tmp/server-db.json" {
		t.Fatalf("file storage path: got %q, want %q", cfg.fileStoragePath, "/tmp/server-db.json")
	}
	if !cfg.restore {
		t.Fatal("restore: got false, want true")
	}
	if cfg.databaseDSN != dsn {
		t.Fatalf("database dsn: got %q, want %q", cfg.databaseDSN, dsn)
	}
}

func TestLoadServerConfigEnvOverridesFlag(t *testing.T) {
	const flagDSN = "host=localhost port=5432 user=flag dbname=metrics sslmode=disable"
	const envDSN = "host=localhost port=5432 user=env dbname=metrics sslmode=disable"

	cfg, err := loadServerConfigWithEnv(
		[]string{
			"-a=127.0.0.1:9000",
			"-i=10",
			"-f=flag-db.json",
			"-r=false",
			"-d=" + flagDSN,
		},
		envLookup(map[string]string{
			addressEnvKey:         "localhost:9090",
			storeIntervalEnvKey:   "1",
			fileStoragePathEnvKey: "env-db.json",
			restoreEnvKey:         "true",
			databaseDSNEnvKey:     envDSN,
		}),
	)
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}

	if cfg.address != "localhost:9090" {
		t.Fatalf("address: got %q, want %q", cfg.address, "localhost:9090")
	}
	if cfg.storeInterval != time.Second {
		t.Fatalf("store interval: got %v, want %v", cfg.storeInterval, time.Second)
	}
	if cfg.fileStoragePath != "env-db.json" {
		t.Fatalf("file storage path: got %q, want %q", cfg.fileStoragePath, "env-db.json")
	}
	if !cfg.restore {
		t.Fatal("restore: got false, want true")
	}
	if cfg.databaseDSN != envDSN {
		t.Fatalf("database dsn: got %q, want %q", cfg.databaseDSN, envDSN)
	}
}

func TestLoadServerConfigUsesFlagsWithoutEnv(t *testing.T) {
	const dsn = "host=localhost port=5432 user=flag dbname=metrics sslmode=disable"

	cfg, err := loadServerConfigWithEnv([]string{
		"-a=127.0.0.1:9000",
		"-i=10",
		"-f=flag-db.json",
		"-r=false",
		"-d=" + dsn,
	}, emptyEnvLookup)
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}

	if cfg.address != "127.0.0.1:9000" {
		t.Fatalf("address: got %q, want %q", cfg.address, "127.0.0.1:9000")
	}
	if cfg.storeInterval != 10*time.Second {
		t.Fatalf("store interval: got %v, want %v", cfg.storeInterval, 10*time.Second)
	}
	if cfg.fileStoragePath != "flag-db.json" {
		t.Fatalf("file storage path: got %q, want %q", cfg.fileStoragePath, "flag-db.json")
	}
	if cfg.restore {
		t.Fatal("restore: got true, want false")
	}
	if cfg.databaseDSN != dsn {
		t.Fatalf("database dsn: got %q, want %q", cfg.databaseDSN, dsn)
	}
}

func TestLoadServerConfigRejectsInvalidEnvInterval(t *testing.T) {
	_, err := loadServerConfigWithEnv(nil, envLookup(map[string]string{
		storeIntervalEnvKey: "bad",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadServerConfigRejectsInvalidEnvRestore(t *testing.T) {
	_, err := loadServerConfigWithEnv(nil, envLookup(map[string]string{
		restoreEnvKey: "bad",
	}))
	if err == nil {
		t.Fatal("expected error")
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
