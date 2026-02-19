package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	flagvalues "github.com/Avgys/go-url-shortener-server/internal/config/flag_values"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/shared"
	httpShared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
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
	redirectAddr    *flagvalues.NetAddress
}

func NewShortifier(stringGenerator StringGenerator, store repository.Repository, redirectAddr *flagvalues.NetAddress) *Shortifier {
	return &Shortifier{stringGenerator: stringGenerator, store: store, redirectAddr: redirectAddr}
}

func (s *Shortifier) ShortifyURL(ctx context.Context, inputURL string, traceLogger *zerolog.Logger) (string, error) {

	if inputURL == "" {
		return "", httpShared.NewError("empty url", http.StatusBadRequest)
	}

	if _, err := shared.GetURL(inputURL, true); err != nil {
		return "", httpShared.NewError("url in wrong format", http.StatusBadRequest)
	}

	trimmedURL := strings.TrimSpace(inputURL)

	shortURL := ""

	var storeErr error

	for range maxStoreRetryCount {
		shortURL = s.stringGenerator.GetRandomString(shortURLMaxLength)
		storeErr = s.store.StoreURL(ctx, trimmedURL, shortURL)

		if storeErr != nil {
			// if collision try again
			if errors.Is(storeErr, repository.ErrCollision) {
				continue
			}

			return "", fmt.Errorf("error saving short url in store, %w", storeErr)
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

func (s *Shortifier) ResolveShortURL(ctx context.Context, inputURL string, traceLogger *zerolog.Logger) (string, error) {

	shortURL := strings.TrimSpace(inputURL)

	url, err := s.store.ResolveShortURL(ctx, shortURL)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		traceLogger.Info().
			Str("repository error", err.Error()).
			Msg("url not found in repository")

		return url, httpShared.NewError("url not found", http.StatusNotFound)
	}

	return url, err
}

func (s *Shortifier) ShortifyBatch(ctx context.Context, request []model.IndexedFullURL, traceLogger *zerolog.Logger) (result model.ShortenBatchResp, err error) {

	if len(request) == 0 {
		return nil, httpShared.NewError("empty url", http.StatusOK)
	}

	for _, url := range request {

		if _, err := shared.GetURL(url.FullURL, true); err != nil {
			return nil, httpShared.NewError("url in wrong format", http.StatusBadRequest)
		}

		url.FullURL = strings.TrimSpace(url.FullURL)
	}

	readyBatch := make(map[string]string, len(request))
	full2shortBatch := make(map[string]string, len(request))
	urlsToInsert := lo.Map(request, func(x model.IndexedFullURL, _ int) string { return x.FullURL })

	for range maxStoreRetryCount {
		clear(full2shortBatch)

		for _, fullURL := range urlsToInsert {
			full2shortBatch[fullURL] = s.stringGenerator.GetRandomString(shortURLMaxLength)
		}

		retryToInsert, alreadyExists, storeErr := s.store.StoreBatch(ctx, full2shortBatch)

		if storeErr != nil {
			return nil, fmt.Errorf("error saving short url in store, %w", storeErr)
		}

		for k, v := range alreadyExists {
			readyBatch[k] = v
		}

		for k, v := range full2shortBatch {
			_, exists := alreadyExists[k]
			if exists || lo.Contains(retryToInsert, k) {
				break
			}

			readyBatch[k] = v
		}

		if len(retryToInsert) == 0 {
			break
		}

		urlsToInsert = retryToInsert
	}

	if err != nil {
		return nil, fmt.Errorf("error saving short url in store, %w", err)
	}

	result = lo.Map(request, func(x model.IndexedFullURL, _ int) model.IndexedShortURL {
		return model.IndexedShortURL{CorrelationId: x.CorrelationId, ShortURL: readyBatch[x.FullURL]}
	})

	return
}
