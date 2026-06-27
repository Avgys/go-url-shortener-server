package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"
)

func useSelfSignedTLS(hosts []string) bool {
	if len(hosts) == 0 {
		return true
	}

	for _, host := range hosts {
		if !isPublicACMEDomain(host) {
			return true
		}
	}

	return false
}

func isPublicACMEDomain(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))

	if host == "localhost" {
		return false
	}

	if ip := net.ParseIP(host); ip != nil {
		return false
	}

	for _, name := range []string{"example.com", "example.net", "example.org"} {
		if host == name {
			return false
		}
	}

	for _, suffix := range []string{
		".example.com", ".example.net", ".example.org",
		".invalid", ".test", ".localhost",
	} {
		if strings.HasSuffix(host, suffix) {
			return false
		}
	}

	return true
}

func generateSelfSignedTLSConfig(hosts []string) (*tls.Config, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate private key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("generate serial: %w", err)
	}

	dnsNames := make([]string, 0, len(hosts))
	ipAddresses := make([]net.IP, 0)

	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			ipAddresses = append(ipAddresses, ip)
			continue
		}
		dnsNames = append(dnsNames, host)
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"go-url-shortener"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		IPAddresses:           ipAddresses,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("load key pair: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}
