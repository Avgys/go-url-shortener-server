package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type JsonConfig struct {
	AppURL             string `json:"server_address,omitempty"`
	RedirectDomain     string `json:"base_url,omitempty"`
	FileStoragePath    string `json:"file_storage_path,omitempty"`
	DBConnectionString string `json:"database_dsn,omitempty"`
	AuditFile          string `json:"audit_file,omitempty"`
	AuditURL           string `json:"audit_url,omitempty"`
	HttpsEnabled       bool   `json:"enable_https,omitempty"`
}

func getJsonConfig(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("shortener-config", flag.ContinueOnError)

	var configPath string
	fs.StringVar(&configPath, "c", "", "config path")
	fs.Parse(args)

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

	defer f.Close()

	var newCfg JsonConfig
	err = json.NewDecoder(f).Decode(&newCfg)

	if err != nil {
		return fmt.Errorf("error decoding config file, %w", err)
	}

	cfg.AppURL.Set(newCfg.AppURL)
	cfg.RedirectDomain.Set(newCfg.RedirectDomain)
	cfg.FileStoragePath = newCfg.FileStoragePath
	cfg.DBConnectionString = newCfg.DBConnectionString
	cfg.AuditFile = newCfg.AuditFile
	cfg.AuditURL = newCfg.AuditURL
	cfg.HttpsEnabled = newCfg.HttpsEnabled

	return nil
}
