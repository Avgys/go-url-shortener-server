package config

import (
	"flag"
	"reflect"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

type Config struct {
	AppURL         NetAddress `env:"SERVER_ADDRESS"`
	RedirectDomain NetAddress `env:"BASE_URL"`
	FileStorage    Filepath   `env:"FILE_STORAGE_PATH"`
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

	log.Info().
		Str("ServerAddr", cfg.AppURL.String()).
		Str("RedirectAddr", cfg.RedirectDomain.String()).
		Str("FileStoragePath", cfg.FileStorage.String())

	return cfg, nil
}

func parseEnv(cfg *Config) error {
	err := env.ParseWithFuncs(cfg, map[reflect.Type]env.ParserFunc{
		reflect.TypeFor[NetAddress](): func(v string) (interface{}, error) {
			netAddress := NetAddress{}
			err := netAddress.Set(v)
			return netAddress, err
		},
		reflect.TypeFor[Filepath](): func(v string) (interface{}, error) {
			filepath := Filepath(v)
			return filepath, nil
		},
	})

	return err
}

func parseFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)

	fs.Var(&cfg.AppURL, "a", "address of HTTP server")
	fs.Var(&cfg.RedirectDomain, "b", "address of redirect")
	fs.Var(&cfg.FileStorage, "f", "file storage name")

	return fs.Parse(args)
}

func getDefaultConfig() *Config {
	cfg := Config{}

	cfg.AppURL = NetAddress{Host: "localhost:8080", SchemeRequired: false}
	cfg.RedirectDomain = NetAddress{Host: "localhost:8080", Scheme: "http", SchemeRequired: true}
	cfg.FileStorage = "../storage.json"

	return &cfg
}
