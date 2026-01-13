package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/shared"
)

const shortURLMaxLength = 8
const maxStoreRetryCount = 20

var (
	ErrInvalidURL  = errors.New("url in wrong format")
	ErrURLNotFound = errors.New("url not found in store")
)

type StringGenerator interface {
	GetRandomString(n int) string
}

type Repository interface {
	StoreURL(url string, urlHash string) error
	ResolveShortURL(shortURL string) (string, bool)
}

type Shortifier struct {
	domain          string
	store           Repository
	stringGenerator StringGenerator
	redirectAddr    *config.NetAddress
}

func NewShortifier(stringGenerator StringGenerator, store Repository, redirectAddr *config.NetAddress) *Shortifier {

	return &Shortifier{stringGenerator: stringGenerator, store: store, redirectAddr: redirectAddr}
}

func (s *Shortifier) ShortifyURL(inputURL string) (string, bool, error) {
	if _, err := shared.GetURL(inputURL, true); err != nil {
		return "", false, ErrInvalidURL
	}

	trimmedURL := strings.TrimSpace(inputURL)

	isStoredNew := false
	shortURL := ""

	for i := 0; !isStoredNew && i < maxStoreRetryCount; i++ {
		shortURL = s.stringGenerator.GetRandomString(shortURLMaxLength)

		isStoredNew = s.store.StoreURL(trimmedURL, shortURL)
	}

	resultURL := ""
	if isStoredNew {
		resultURL, _ = url.JoinPath(s.redirectAddr.String(), shortURL)
	}

	return resultURL, isStoredNew, nil
}

func (s *Shortifier) ResolveShortURL(shortURL strifng) (string, error) {

	shortURL = strings.TrimSpace(shortURL)

	url, ok := s.store.ResolveShortURL(shortURL)

	if !ok {
		return "", fmt.Errorf("%s %w", shortURL, ErrURLNotFound)
	}

	return url, nil
}
