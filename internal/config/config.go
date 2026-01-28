package config

import (
	"flag"
	"log"
	"reflect"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	AppURL         NetAddress `env:"SERVER_ADDRESS"`
	RedirectDomain NetAddress `env:"BASE_URL"`
}

func GetConfig(args []string) (*Config, error) {
	cfg := getDefaultConfig()

	err := parseFlags(cfg, args)

	if err != nil {
		return nil, err
	}

	err = parseEnv(cfg)

	if err != nil {
		return nil, err
	}

	log.Printf("ServerAddr: %s, RedirectAddr: %s", cfg.AppURL.String(), cfg.RedirectDomain.String())

	return cfg, nil
}

func parseEnv(cfg *Config) error {
	err := env.ParseWithFuncs(cfg, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(NetAddress{}): func(v string) (interface{}, error) {
			val, err := NewNetAddress(v)

			if err != nil {
				return nil, err
			}

			return *val, nil
		}})

	if err != nil {
		return err
	}

	return nil
}

func parseFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)

	fs.Var(&cfg.AppURL, "a", "address of HTTP server")
	fs.Var(&cfg.RedirectDomain, "b", "address of redirect")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return nil
}

func getDefaultConfig() *Config {
	cfg := Config{}

	cfg.AppURL = NetAddress{Host: "localhost:8080", SchemeRequired: false}
	cfg.RedirectDomain = NetAddress{Host: "localhost:8080", Scheme: "http", SchemeRequired: true}

	return &cfg
}
