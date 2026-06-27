package config

import (
	"flag"
	"fmt"
	"reflect"

	flagvalues "go-url-shortener/internal/config/flag_values"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
)

type Config struct {
	AppURL             flagvalues.NetAddress `env:"SERVER_ADDRESS"`
	RedirectDomain     flagvalues.NetAddress `env:"BASE_URL"`
	FileStoragePath    string                `env:"FILE_STORAGE_PATH"`
	DBConnectionString string                `env:"DATABASE_DSN"`
	AuditFile          string                `env:"AUDIT_FILE"`
	AuditURL           string                `env:"AUDIT_URL"`
	HttpsEnabled       bool                  `env:"ENABLE_HTTPS"`
}

func GetConfig(args []string, traceLogger *zerolog.Logger) (*Config, error) {
	cfg := getDefaultConfig()

	err := parseFlags(cfg, args)

	if err != nil {
		return nil, fmt.Errorf("error parsing flags for config, %w", err)
	}

	err = parseEnv(cfg)

	if err != nil {
		return nil, fmt.Errorf("error parsing env variables for config, %w", err)
	}

	traceLogger.Info().
		Str("ServerAddr", cfg.AppURL.String()).
		Str("RedirectAddr", cfg.RedirectDomain.String()).
		Str("FileStoragePath", cfg.FileStoragePath).
		Str("DBConnectionString", cfg.DBConnectionString).
		Str("AuditFile", cfg.AuditFile).
		Str("AuditURL", cfg.AuditURL).
		Bool("HttpsEnabled", cfg.HttpsEnabled).
		Send()

	return cfg, nil
}

func parseEnv(cfg *Config) error {
	err := env.ParseWithOptions(cfg, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeFor[flagvalues.NetAddress](): func(v string) (interface{}, error) {
				netAddress := flagvalues.NetAddress{}
				err := netAddress.Set(v)
				return netAddress, err
			},
		}})

	return err
}

func parseFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)

	fs.Var(&cfg.AppURL, "a", "address of HTTP server")
	fs.Var(&cfg.RedirectDomain, "b", "address of redirect")
	fs.StringVar(&cfg.FileStoragePath, "f", "", "file storage name")
	fs.StringVar(&cfg.DBConnectionString, "d", "", "db connection string url")
	fs.StringVar(&cfg.AuditFile, "audit-file", "", "path to audit log file (empty disables file audit)")
	fs.StringVar(&cfg.AuditURL, "audit-url", "", "remote audit receiver URL (empty disables remote audit)")
	fs.BoolVar(&cfg.HttpsEnabled, "s", false, "enable https")

	return fs.Parse(args)
}

func getDefaultConfig() *Config {
	cfg := Config{}

	cfg.AppURL = flagvalues.NetAddress{Host: "localhost:8080", SchemeRequired: false}
	cfg.RedirectDomain = flagvalues.NetAddress{Host: "localhost:8080", Scheme: "http", SchemeRequired: true}
	cfg.FileStoragePath = ""
	cfg.DBConnectionString = ""
	cfg.AuditFile = ""
	cfg.AuditURL = ""
	cfg.HttpsEnabled = false

	return &cfg
}
