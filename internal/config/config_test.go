package config

import (
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

type ConfigSuite struct {
	suite.Suite
}

func (s *ConfigSuite) clearEnv(key string) {
	value, wasSet := os.LookupEnv(key)
	s.Require().NoError(os.Unsetenv(key))

	s.T().Cleanup(func() {
		if wasSet {
			_ = os.Setenv(key, value)
			return
		}
		_ = os.Unsetenv(key)
	})
}

func (s *ConfigSuite) TestParseFlags() {
	cfg := &Config{}

	err := parseFlags(cfg, []string{
		"-a", "localhost:8080",
		"-d", "db",
		"-b", "http://localhost:8080",
		"-f", "/tmp/storage",
		"-audit-file", "/var/log/audit.log",
		"-audit-url", "http://audit.example.com/events",
	})
	s.Require().NoError(err)

	s.Equal("localhost:8080", cfg.AppURL.Host)
	s.Equal("db", cfg.DBConnectionString)
	s.Equal("http://localhost:8080", cfg.RedirectDomain.String())
	s.Equal("/tmp/storage", cfg.FileStoragePath)
	s.Equal("/var/log/audit.log", cfg.AuditFile)
	s.Equal("http://audit.example.com/events", cfg.AuditURL)
}

func (s *ConfigSuite) TestParseFlags_AuditDisabledByDefault() {
	cfg := &Config{}

	err := parseFlags(cfg, []string{"-a", "localhost:8080"})
	s.Require().NoError(err)

	s.Empty(cfg.AuditFile)
	s.Empty(cfg.AuditURL)
}

func (s *ConfigSuite) TestParseFlags_UnknownFlag() {
	cfg := &Config{}

	err := parseFlags(cfg, []string{"-unknown"})
	s.Error(err)
}

func (s *ConfigSuite) TestGetConfig_EnvOverridesFlags() {
	s.T().Setenv("SERVER_ADDRESS", "env:8081")
	s.T().Setenv("DATABASE_DSN", "envdb")
	s.T().Setenv("BASE_URL", "http://env-base")
	s.T().Setenv("FILE_STORAGE_PATH", "/env/storage")
	s.T().Setenv("AUDIT_FILE", "/env/audit.log")
	s.T().Setenv("AUDIT_URL", "http://env-audit.example.com")

	logger := zerolog.New(io.Discard)
	cfg, err := GetConfig([]string{"-a", "flag:8080", "-d", "flagdb", "-b", "http://flag-base", "-f", "/flag/storage"}, &logger)
	s.Require().NoError(err)

	s.Equal("env:8081", cfg.AppURL.Host)
	s.Equal("envdb", cfg.DBConnectionString)
	s.Equal("http://env-base", cfg.RedirectDomain.String())
	s.Equal("/env/storage", cfg.FileStoragePath)
	s.Equal("/env/audit.log", cfg.AuditFile)
	s.Equal("http://env-audit.example.com", cfg.AuditURL)
}

func (s *ConfigSuite) TestGetConfig_FlagsOnly() {
	s.clearEnv("SERVER_ADDRESS")
	s.clearEnv("DATABASE_DSN")
	s.clearEnv("BASE_URL")
	s.clearEnv("FILE_STORAGE_PATH")
	s.clearEnv("AUDIT_FILE")
	s.clearEnv("AUDIT_URL")

	logger := zerolog.New(io.Discard)
	cfg, err := GetConfig([]string{"-a", "flag:8080", "-d", "flagdb", "-b", "http://flag-base", "-f", "/flag/storage"}, &logger)
	s.Require().NoError(err)

	s.Equal("flag:8080", cfg.AppURL.Host)
	s.Equal("flagdb", cfg.DBConnectionString)
	s.Equal("http://flag-base", cfg.RedirectDomain.String())
	s.Equal("/flag/storage", cfg.FileStoragePath)
}
