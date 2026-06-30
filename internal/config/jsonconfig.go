package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
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
	configArgs := make([]string, 0, 2)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-c" {
			configArgs = append(configArgs, arg)

			if i+1 < len(args) {
				configArgs = append(configArgs, args[i+1])
				i++
			}

			continue
		}

		if strings.HasPrefix(arg, "-c=") {
			configArgs = append(configArgs, arg)
		}
	}

	if err := fs.Parse(configArgs); err != nil {
		return fmt.Errorf("error parsing config path flag, %w", err)
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

	var newCfg JsonConfig
	err = json.NewDecoder(f).Decode(&newCfg)

	if err != nil {
		return fmt.Errorf("error decoding config file, %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing config file, %w", err)
	}

	if err := cfg.AppURL.Set(newCfg.AppURL); err != nil {
		return fmt.Errorf("error parsing server_address from config file, %w", err)
	}

	if err := cfg.RedirectDomain.Set(newCfg.RedirectDomain); err != nil {
		return fmt.Errorf("error parsing base_url from config file, %w", err)
	}

	cfg.FileStoragePath = newCfg.FileStoragePath
	cfg.DBConnectionString = newCfg.DBConnectionString
	cfg.AuditFile = newCfg.AuditFile
	cfg.AuditURL = newCfg.AuditURL
	cfg.HttpsEnabled = newCfg.HttpsEnabled

	return nil
}
