package main

import "os"

const addressEnvKey = "ADDRESS"

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

	return cfg, nil
}
