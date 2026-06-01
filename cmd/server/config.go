package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	addressEnvKey         = "ADDRESS"
	storeIntervalEnvKey   = "STORE_INTERVAL"
	fileStoragePathEnvKey = "FILE_STORAGE_PATH"
	restoreEnvKey         = "RESTORE"
	databaseDSNEnvKey     = "DATABASE_DSN"
)

func loadServerConfig(args []string) (serverConfig, error) {
	return loadServerConfigWithEnv(args, os.LookupEnv)
}

func loadServerConfigWithEnv(args []string, lookupEnv func(string) (string, bool)) (serverConfig, error) {
	cfg, err := parseServerFlags(args)
	if err != nil {
		return serverConfig{}, err
	}

	if value, ok := lookupEnv(addressEnvKey); ok {
		cfg.address = value
	}

	if value, ok := lookupEnv(storeIntervalEnvKey); ok {
		seconds, err := parseEnvSeconds(storeIntervalEnvKey, value)
		if err != nil {
			return serverConfig{}, err
		}
		cfg.storeInterval = time.Duration(seconds) * time.Second
	}

	if value, ok := lookupEnv(fileStoragePathEnvKey); ok {
		cfg.fileStoragePath = value
		cfg.fileStoragePathProvided = true
	}

	if value, ok := lookupEnv(restoreEnvKey); ok {
		restore, err := strconv.ParseBool(value)
		if err != nil {
			return serverConfig{}, fmt.Errorf("%s must be a boolean value", restoreEnvKey)
		}
		cfg.restore = restore
	}

	if value, ok := lookupEnv(databaseDSNEnvKey); ok {
		cfg.databaseDSN = value
		cfg.databaseDSNProvided = true
	}

	cfg.resolveStorageMode()

	return cfg, nil
}

func parseEnvSeconds(key string, value string) (int, error) {
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer number of seconds", key)
	}
	if seconds < 0 {
		return 0, fmt.Errorf("%s must be non-negative", key)
	}

	return seconds, nil
}
