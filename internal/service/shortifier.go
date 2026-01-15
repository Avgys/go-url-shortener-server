package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/shared"
)

const shortURLMaxLength = 8
const maxStoreRetryCount = 20

var (
	ErrInvalidURL = errors.New("url in wrong format")
	ErrCollision  = errors.New("could not find free space to store url")
	ErrNotFound   = errors.New("url not found")
)

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
	if _, err := shared.GetURL(inputURL, true); err != nil {
		return "", ErrInvalidURL
	}

	trimmedURL := strings.TrimSpace(inputURL)

	shortURL := ""

	for range maxStoreRetryCount {
		shortURL = s.stringGenerator.GetRandomString(shortURLMaxLength)

		if err := s.store.StoreURL(trimmedURL, shortURL); err != nil {
			if !errors.Is(err, repository.ErrCollision) {
				return "", ErrCollision
			}
		} else {
			break
		}
	}

	resultURL, err := url.JoinPath(s.redirectAddr.String(), shortURL)

	if err != nil {
		return "", err
	}

	return resultURL, nil
}

func (s *Shortifier) ResolveShortURL(shortURL string) (string, error) {

	shortURL = strings.TrimSpace(shortURL)

	url, err := s.store.ResolveShortURL(shortURL)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", fmt.Errorf("%s %w, inner error: %w", shortURL, ErrNotFound, err)
		} else {
			return "", err
		}
	}

	return url, nil
}
