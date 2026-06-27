package server

import (
	"testing"

	flagvalues "go-url-shortener/internal/config/flag_values"
	"go-url-shortener/internal/config"
)

func TestAutocertHosts(t *testing.T) {
	cfg := &config.Config{
		AppURL:         flagvalues.NetAddress{Host: "short.example.com:443"},
		RedirectDomain: flagvalues.NetAddress{Host: "short.example.com:443"},
	}

	got := autocertHosts(cfg)
	if len(got) != 1 || got[0] != "short.example.com" {
		t.Fatalf("got %v, want [short.example.com]", got)
	}
}

func TestHostnameFromAddr(t *testing.T) {
	if got := hostnameFromAddr("localhost:8080"); got != "localhost" {
		t.Fatalf("got %q", got)
	}
}
