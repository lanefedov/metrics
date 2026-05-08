package main

import (
	"flag"
	"fmt"
	"io"
)

const defaultListenAddress = "localhost:8080"

type serverConfig struct {
	address string
}

func parseServerFlags(args []string) (serverConfig, error) {
	cfg := serverConfig{
		address: defaultListenAddress,
	}

	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.address, "a", defaultListenAddress, "HTTP server address")

	if err := fs.Parse(args); err != nil {
		return serverConfig{}, err
	}

	if fs.NArg() > 0 {
		return serverConfig{}, fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	return cfg, nil
}
