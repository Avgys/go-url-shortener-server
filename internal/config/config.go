package config

import (
	"fmt"

	flagvalues "go-url-shortener/internal/config/flag_values"

	"github.com/rs/zerolog"
)

type Config struct {
	AppURL             flagvalues.NetAddress `env:"SERVER_ADDRESS" json:"server_address,omitempty"`
	RedirectDomain     flagvalues.NetAddress `env:"BASE_URL" json:"base_url,omitempty"`
	FileStoragePath    string                `env:"FILE_STORAGE_PATH" json:"file_storage_path,omitempty"`
	DBConnectionString string                `env:"DATABASE_DSN" json:"database_dsn,omitempty"`
	AuditFile          string                `env:"AUDIT_FILE" json:"audit_file,omitempty"`
	AuditURL           string                `env:"AUDIT_URL" json:"audit_url,omitempty"`
	HttpsEnabled       bool                  `env:"ENABLE_HTTPS" json:"enable_https,omitempty"`
}

func GetConfig(args []string, traceLogger *zerolog.Logger) (*Config, error) {
	cfg := getDefaultConfig()

	if err := getJsonConfig(cfg, args); err != nil {
		return nil, fmt.Errorf("error parsing config file, %w", err)
	}

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
