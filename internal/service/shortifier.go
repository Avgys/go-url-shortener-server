package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/shared"
	"github.com/rs/zerolog"
)

const shortURLMaxLength = 8
const maxStoreRetryCount = 20

var (
	ErrCollision = errors.New("could not find free space to store url")
)

type StringGenerator interface {
	GetRandomString(n int) string
}

type Shortifier struct {
	domain          string
	store           repository.Repository
	stringGenerator StringGenerator
	redirectAddr    *config.NetAddress
}

func NewShortifier(stringGenerator StringGenerator, store repository.Repository, redirectAddr *config.NetAddress) *Shortifier {
	return &Shortifier{stringGenerator: stringGenerator, store: store, redirectAddr: redirectAddr}
}

func (s *Shortifier) ShortifyURL(inputURL string, traceLogger *zerolog.Logger) (string, error) {

	if inputURL == "" {
		return "", logger.NewError("empty url", http.StatusBadRequest)
	}

	if _, err := shared.GetURL(inputURL, true); err != nil {
		return "", logger.NewError("url in wrong format", http.StatusBadRequest)
	}

	trimmedURL := strings.TrimSpace(inputURL)

	shortURL := ""

	var storeErr error

	for range maxStoreRetryCount {
		shortURL = s.stringGenerator.GetRandomString(shortURLMaxLength)
		storeErr = s.store.StoreURL(trimmedURL, shortURL)

		if storeErr != nil {
			// if collision try again
			if errors.Is(storeErr, repository.ErrCollision) {
				continue
			}

			return "", fmt.Errorf("Error saving short url in store, %w", storeErr)
		}

		// if no errors, then value stored successfuly
		break
	}

	// tries exceed retry count
	if storeErr != nil && errors.Is(storeErr, repository.ErrCollision) {
		return "", ErrCollision
	}

	return url.JoinPath(s.redirectAddr.String(), shortURL)
}

func (s *Shortifier) ResolveShortURL(inputURL string, traceLogger *zerolog.Logger) (string, error) {

	shortURL := strings.TrimSpace(inputURL)

	url, err := s.store.ResolveShortURL(shortURL)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		traceLogger.Info().
			Str("repository error", err.Error()).
			Msg("Url not found in repository")

		return url, logger.NewError("url not found", http.StatusNotFound)
	}

	return url, err
}
