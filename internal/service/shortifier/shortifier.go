package shortifier

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrInvalidURL  = errors.New("url in wrong format")
	ErrURLNotFound = errors.New("url not found in store")
)

type Hasher interface {
	GetHash(input string) string
}

type Repository interface {
	StoreURL(url string, urlHash string) bool
	ResolveShortURL(shortURL string) (string, bool)
}

type Shortifier struct {
	domain   string
	store    Repository
	hashFunc Hasher
}

func NewShortifier(hashFunc Hasher, store Repository, domain string) *Shortifier {
	return &Shortifier{hashFunc: hashFunc, store: store, domain: domain}
}

func (s *Shortifier) ShortifyURL(url string) (string, bool, error) {
	if !isValidURL(url) {
		return "", false, ErrInvalidURL
	}

	url = strings.TrimSpace(url)

	shortURL := s.hashFunc.GetHash(url)

	isCreated := s.store.StoreURL(url, shortURL)
	readyToUseURL := s.domain + "/" + shortURL
	return readyToUseURL, isCreated, nil
}

func (s *Shortifier) ResolveShortURL(shortURL string) (string, error) {

	shortURL = strings.TrimSpace(shortURL)

	url, ok := s.store.ResolveShortURL(shortURL)

	if !ok {
		return "", fmt.Errorf("%s %w", shortURL, ErrURLNotFound)
	}

	return url, nil
}

// isValidURL reports whether s is a syntactically valid absolute HTTP/HTTPS URL.
// Requirements:
// - Must parse via url.ParseRequestURI
// - Scheme must be http or https
// - Host must be non-empty
func isValidURL(s string) bool {
	if s == "" {
		return false
	}
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}
