package server

import "testing"

func TestIsPublicACMEDomain(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"short.example.com", false},
		{"example.com", false},
		{"localhost", false},
		{"127.0.0.1", false},
		{"mysite.ru", true},
	}

	for _, tt := range tests {
		if got := isPublicACMEDomain(tt.host); got != tt.want {
			t.Errorf("isPublicACMEDomain(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}

func TestUseSelfSignedTLS(t *testing.T) {
	if !useSelfSignedTLS([]string{"short.example.com"}) {
		t.Fatal("expected self-signed for example.com domain")
	}
	if useSelfSignedTLS([]string{"short.mysite.ru"}) {
		t.Fatal("expected public ACME for real domain")
	}
}
