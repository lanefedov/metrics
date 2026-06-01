package main

import (
	"log"
	"os"

	"github.com/lanefedov/metrics/internal/agent"
)

func main() {
	cfg, err := loadAgentConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	collector := agent.NewCollector(nil, nil)
	reporter := agent.NewReporter(cfg.ServerAddress, nil)
	app := agent.New(cfg, collector, reporter, log.Default())

	app.Run()
}
