package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lanefedov/metrics/internal/agent"
)

func main() {
	cfg, err := loadAgentConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	collector := agent.NewCollector(nil, nil)
	gopsCollector := agent.NewGopsutilCollector()
	reporter := agent.NewReporter(cfg.ServerAddress, cfg.Key, nil)
	app := agent.New(cfg, collector, gopsCollector, reporter, log.Default())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app.Run(ctx)
}
