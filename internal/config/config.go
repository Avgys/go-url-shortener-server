package config

import (
	"flag"
)

type Config struct {
	AppURL         NetAddress
	RedirectDomain NetAddress
}

func GetConfig(args []string) (*Config, error) {
	cfg := getDefaultConfig()

	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)

	fs.Var(&cfg.AppURL, "a", "address of HTTP server")
	fs.Var(&cfg.RedirectDomain, "b", "address of redirect")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getDefaultConfig() *Config {
	cfg := Config{}

	cfg.AppURL = NetAddress{Host: "localhost:8080", SchemeRequired: false}
	cfg.RedirectDomain = NetAddress{Host: "localhost:8080", Scheme: "http", SchemeRequired: true}

	return &cfg
}
