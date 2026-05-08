package main

import (
	"log"

	"github.com/lanefedov/metrics/internal/agent"
)

func main() {
	cfg := agent.DefaultConfig()
	collector := agent.NewCollector()
	reporter := agent.NewReporter(cfg.ServerAddress, nil)
	app := agent.New(cfg, collector, reporter, log.Default())

	app.Run()
}
