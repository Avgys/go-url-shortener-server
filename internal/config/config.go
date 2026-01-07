package config

import (
	"flag"
	"net/url"
)

type Config struct {
	URL url.URL
}

func GetConfig() *Config {
	cfg := Config{}

	url := url.URL{}

	flag.StringVar(&(url.Host), "host", "localhost:8080", "address of HTTP server")
	flag.StringVar(&(url.Scheme), "scheme", "http", "address of HTTP server")
	cfg.URL = url

	flag.Parse()
	return &cfg
}

func GetDefaultConfig() *Config {
	cfg := Config{
		URL: url.URL{Host: "localhost:8080", Scheme: "http"},
	}

	return &cfg
}
