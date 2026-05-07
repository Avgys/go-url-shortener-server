package config

import (
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearEnv(t *testing.T, key string) {
	t.Helper()

	value, wasSet := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))

	t.Cleanup(func() {
		if wasSet {
			_ = os.Setenv(key, value)
			return
		}
		_ = os.Unsetenv(key)
	})
}

func TestParseFlags(t *testing.T) {
	cfg := &Config{}

	err := parseFlags(cfg, []string{"-a", "localhost:8080", "-d", "db", "-r", "http://accrual"})
	require.NoError(t, err)

	assert.Equal(t, "localhost:8080", cfg.AppAddr)
	assert.Equal(t, "db", cfg.DBConnectionString)
	assert.Equal(t, "http://accrual", cfg.AccrualSystemAddr)
}

func TestParseFlags_UnknownFlag(t *testing.T) {
	cfg := &Config{}

	err := parseFlags(cfg, []string{"-unknown"})
	require.Error(t, err)
}

func TestGetConfig_FlagsOverrideEnv(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "env:8081")
	t.Setenv("DATABASE_URI", "envdb")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://env-accrual")

	logger := zerolog.New(io.Discard)
	cfg, err := GetConfig([]string{"-a", "flag:8080", "-d", "flagdb", "-r", "http://flag-accrual"}, &logger)
	require.NoError(t, err)

	assert.Equal(t, "flag:8080", cfg.AppAddr)
	assert.Equal(t, "flagdb", cfg.DBConnectionString)
	assert.Equal(t, "http://flag-accrual", cfg.AccrualSystemAddr)
}

func TestGetConfig_FlagsOnly(t *testing.T) {
	clearEnv(t, "RUN_ADDRESS")
	clearEnv(t, "DATABASE_URI")
	clearEnv(t, "ACCRUAL_SYSTEM_ADDRESS")

	logger := zerolog.New(io.Discard)
	cfg, err := GetConfig([]string{"-a", "flag:8080", "-d", "flagdb", "-r", "http://flag-accrual"}, &logger)
	require.NoError(t, err)

	assert.Equal(t, "flag:8080", cfg.AppAddr)
	assert.Equal(t, "flagdb", cfg.DBConnectionString)
	assert.Equal(t, "http://flag-accrual", cfg.AccrualSystemAddr)
}
