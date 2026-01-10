package shortifier

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/shared"
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
	domain       string
	store        Repository
	hashFunc     Hasher
	redirectAddr *config.NetAddress
}

func NewShortifier(hashFunc Hasher, store Repository, redirectAddr *config.NetAddress) *Shortifier {

	return &Shortifier{hashFunc: hashFunc, store: store, redirectAddr: redirectAddr}
}

func (s *Shortifier) ShortifyURL(url string) (string, bool, error) {
	if _, err := shared.GetURL(url, true); err != nil {
		return "", false, ErrInvalidURL
	}

	url = strings.TrimSpace(url)

	shortURL := s.hashFunc.GetHash(url)

	isCreated := s.store.StoreURL(url, shortURL)

	resultURL := fmt.Sprintf("%s/%s", s.redirectAddr.String(), shortURL)

	return resultURL, isCreated, nil
}

func (s *Shortifier) ResolveShortURL(shortURL string) (string, error) {

	shortURL = strings.TrimSpace(shortURL)

	url, ok := s.store.ResolveShortURL(shortURL)

	if !ok {
		return "", fmt.Errorf("%s %w", shortURL, ErrURLNotFound)
	}

	return url, nil
}
