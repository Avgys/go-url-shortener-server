package config

import (
	"flag"
)

type Config struct {
	AppURL         NetAddress
	RedirectDomain NetAddress
}

func GetConfig() *Config {
	cfg := getDefaultConfig()

	flag.Var(&cfg.AppURL, "a", "address of HTTP server")
	flag.Var(&cfg.RedirectDomain, "b", "address of redirect")

	flag.Parse()
	return cfg
}

func getDefaultConfig() *Config {
	cfg := Config{}

	cfg.AppURL = NetAddress{Host: "", SchemeRequired: false}
	cfg.RedirectDomain = NetAddress{Host: "localhost:8080", Scheme: "http", SchemeRequired: true}

	return &cfg
}
