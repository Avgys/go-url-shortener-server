package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
)

type JSONConfig struct {
	AppURL             string `json:"server_address,omitempty"`
	RedirectDomain     string `json:"base_url,omitempty"`
	FileStoragePath    string `json:"file_storage_path,omitempty"`
	DBConnectionString string `json:"database_dsn,omitempty"`
	AuditFile          string `json:"audit_file,omitempty"`
	AuditURL           string `json:"audit_url,omitempty"`
	HttpsEnabled       *bool  `json:"enable_https,omitempty"`
	TrustedSubnet      string `json:"trusted_subnet,omitempty"`
	GRPCPort           *int   `json:"grpc_port,omitempty"`
}

func setStringIfNotEmpty(dst *string, src string) {
	if src != "" {
		*dst = src
	}
}

var jsonConfigName flagName = flagName{short: "-c", long: "-c="}

func parseJSONConfig(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("shortener-config", flag.ContinueOnError)

	args = filterFlags(args, []flagName{jsonConfigName})

	var configPath string
	fs.StringVar(&configPath, "c", "", "config path")

	err := fs.Parse(args)
	if err != nil && errors.Is(err, flag.ErrHelp) {
		return fmt.Errorf("error parsing flags, %w", err)
	}

	value, ok := os.LookupEnv("CONFIG")

	if ok {
		configPath = value
	}

	if configPath == "" {
		return nil
	}

	f, err := os.Open(configPath)

	if err != nil {
		return fmt.Errorf("error opening config file, %w", err)
	}

	var newCfg JSONConfig
	err = json.NewDecoder(f).Decode(&newCfg)

	if err != nil {
		return fmt.Errorf("error decoding config file, %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing config file, %w", err)
	}

	if newCfg.AppURL != "" {
		if err := cfg.AppURL.Set(newCfg.AppURL); err != nil {
			return fmt.Errorf("error parsing server_address from config file, %w", err)
		}
	}

	if newCfg.RedirectDomain != "" {
		if err := cfg.RedirectDomain.Set(newCfg.RedirectDomain); err != nil {
			return fmt.Errorf("error parsing base_url from config file, %w", err)
		}
	}

	setStringIfNotEmpty(&cfg.FileStoragePath, newCfg.FileStoragePath)
	setStringIfNotEmpty(&cfg.DBConnectionString, newCfg.DBConnectionString)
	setStringIfNotEmpty(&cfg.AuditFile, newCfg.AuditFile)
	setStringIfNotEmpty(&cfg.AuditURL, newCfg.AuditURL)
	setStringIfNotEmpty(&cfg.TrustedSubnet, newCfg.TrustedSubnet)

	if newCfg.HttpsEnabled != nil {
		cfg.HttpsEnabled = *newCfg.HttpsEnabled
	}

	if newCfg.GRPCPort != nil {
		cfg.GRPCPort = *newCfg.GRPCPort
	}

	return nil
}
