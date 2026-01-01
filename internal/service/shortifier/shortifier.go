package shortifier

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/repository"
)

type Hasher interface {
	GetHash(input string) string
}

type Repository interface {
	StoreUrl(url string, urlHash string) bool
	ResolveShortUrl(shortUrl string) (string, error)
}

type Shortifier struct {
	domain   string
	store    Repository
	hashFunc Hasher
}

func NewShortifier(hashFunc Hasher, store Repository, domain string) *Shortifier {

	return &Shortifier{hashFunc: hashFunc, store: store, domain: domain}
}

func (s *Shortifier) ShortifyUrl(url string) (string, bool, error) {
	if !isValidURL(url) {
		return "", false, errors.New("url in wrong format")
	}

	url = strings.TrimSpace(url)

	shortUrl := s.hashFunc.GetHash(url)

	isCreated := s.store.StoreUrl(url, shortUrl)
	readyToUseUrl := s.domain + "/" + shortUrl
	return readyToUseUrl, isCreated, nil
}

func (s *Shortifier) ResolveShortUrl(shortUrl string) (string, error) {

	shortUrl = strings.TrimSpace(shortUrl)

	url, err := s.store.ResolveShortUrl(shortUrl)

	if err != nil {
		if errors.Is(err, repository.StoreNotFoundErr) {
			return "", fmt.Errorf("Counldn't find url: %w", err)
		}
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
