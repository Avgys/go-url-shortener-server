package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/shared"
	"github.com/Avgys/go-url-shortener-server/internal/shared/common_errors"
)

const shortURLMaxLength = 8
const maxStoreRetryCount = 20

type StringGenerator interface {
	GetRandomString(n int) string
}

type Repository interface {
	StoreURL(url string, urlHash string) error
	ResolveShortURL(shortURL string) (string, error)
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

func (s *Shortifier) ShortifyURL(inputURL string) (string, error) {

	if inputURL == "" {
		return "", common_errors.ErrEmptyURL
	}

	if _, err := shared.GetURL(inputURL, true); err != nil {
		return "", common_errors.ErrInvalidURL
	}

	trimmedURL := strings.TrimSpace(inputURL)

	shortURL := ""

	var storeErr error

	for range maxStoreRetryCount {
		shortURL = s.stringGenerator.GetRandomString(shortURLMaxLength)
		storeErr = s.store.StoreURL(trimmedURL, shortURL)

		if storeErr == nil {
			break
		} else if !errors.Is(storeErr, repository.ErrCollision) {
			return "", storeErr
		}
	}

	if storeErr != nil {
		return "", common_errors.ErrCollision
	}

	return url.JoinPath(s.redirectAddr.String(), shortURL)
}

func (s *Shortifier) ResolveShortURL(inputURL string) (string, error) {

	shortURL := strings.TrimSpace(inputURL)

	url, err := s.store.ResolveShortURL(shortURL)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return url, fmt.Errorf("%s %w, inner error: %w", shortURL, common_errors.ErrNotFound, err)
		}
	}

	return url, err
}
