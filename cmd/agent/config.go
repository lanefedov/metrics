package main

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/lanefedov/metrics/internal/agent"
)

type agentEnvConfig struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

func loadAgentConfig(args []string) (agent.Config, error) {
	cfg, err := parseAgentFlags(args)
	if err != nil {
		return agent.Config{}, err
	}

	envCfg := agentEnvConfig{
		ServerAddress:  cfg.ServerAddress,
		ReportInterval: int(cfg.ReportInterval / time.Second),
		PollInterval:   int(cfg.PollInterval / time.Second),
		Key:            cfg.Key,
		RateLimit:      cfg.RateLimit,
	}

	if err := cleanenv.ReadEnv(&envCfg); err != nil {
		return agent.Config{}, err
	}

	cfg.ServerAddress = envCfg.ServerAddress
	cfg.ReportInterval = time.Duration(envCfg.ReportInterval) * time.Second
	cfg.PollInterval = time.Duration(envCfg.PollInterval) * time.Second
	cfg.Key = envCfg.Key
	cfg.RateLimit = envCfg.RateLimit

	return cfg, nil
}
