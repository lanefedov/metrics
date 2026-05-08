package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/lanefedov/metrics/internal/agent"
)

func parseAgentFlags(args []string) (agent.Config, error) {
	cfg := agent.DefaultConfig()

	var reportSeconds int
	var pollSeconds int

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	fs.IntVar(&reportSeconds, "r", int(cfg.ReportInterval/time.Second), "report interval in seconds")
	fs.IntVar(&pollSeconds, "p", int(cfg.PollInterval/time.Second), "poll interval in seconds")

	if err := fs.Parse(args); err != nil {
		return agent.Config{}, err
	}

	if fs.NArg() > 0 {
		return agent.Config{}, fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	cfg.ReportInterval = time.Duration(reportSeconds) * time.Second
	cfg.PollInterval = time.Duration(pollSeconds) * time.Second

	return cfg, nil
}
