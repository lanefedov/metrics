package main

import (
	"flag"
	"fmt"
	"io"
	"time"
)

const (
	defaultListenAddress = "localhost:8080"
	defaultStoreInterval = 300 * time.Second
	defaultStoreFilePath = "metrics-db.json"
)

type serverConfig struct {
	address         string
	storeInterval   time.Duration
	fileStoragePath string
	restore         bool
	databaseDSN     string
}

func parseServerFlags(args []string) (serverConfig, error) {
	cfg := serverConfig{
		address:         defaultListenAddress,
		storeInterval:   defaultStoreInterval,
		fileStoragePath: defaultStoreFilePath,
		restore:         true,
	}

	var storeIntervalSeconds int

	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.address, "a", defaultListenAddress, "HTTP server address")
	fs.IntVar(&storeIntervalSeconds, "i", int(defaultStoreInterval/time.Second), "store interval in seconds")
	fs.StringVar(&cfg.fileStoragePath, "f", defaultStoreFilePath, "file storage path")
	fs.BoolVar(&cfg.restore, "r", true, "restore metrics from file on startup")
	fs.StringVar(&cfg.databaseDSN, "d", "", "database data source name")

	if err := fs.Parse(args); err != nil {
		return serverConfig{}, err
	}

	if fs.NArg() > 0 {
		return serverConfig{}, fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	if storeIntervalSeconds < 0 {
		return serverConfig{}, fmt.Errorf("store interval must be non-negative")
	}

	cfg.storeInterval = time.Duration(storeIntervalSeconds) * time.Second

	return cfg, nil
}
