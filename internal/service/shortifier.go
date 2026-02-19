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
	ErrConflict  = errors.New("url already stored")
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

type storeInfo struct {
	shortURL string
	new      bool
}

func NewShortifier(stringGenerator StringGenerator, store repository.Repository, redirectAddr *flagvalues.NetAddress) *Shortifier {
	return &Shortifier{stringGenerator: stringGenerator, store: store, redirectAddr: redirectAddr}
}

func (s *Shortifier) ShortifyURL(ctx context.Context, inputURL string, traceLogger *zerolog.Logger) (*model.IndexedShortURL, error) {

	if inputURL == "" {
		return nil, httpShared.NewError("empty url", http.StatusBadRequest)
	}

	if _, err := shared.GetURL(inputURL, true); err != nil {
		return nil, httpShared.NewError("url in wrong format", http.StatusBadRequest)
	}

	result, storeErr := s.ShortifyBatch(ctx, []model.IndexedFullURL{{FullURL: inputURL, CorrelationID: "NO_ID"}}, traceLogger)

	// tries exceed retry count
	if storeErr != nil && errors.Is(storeErr, repository.ErrCollision) {
		return nil, ErrCollision
	}

	shortURL := result[0]

	return &shortURL, nil
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

	readyBatch := make(map[string]storeInfo, len(request))
	full2shortBatch := make(map[string]string, len(request))
	urlsToInsert := lo.Map(request, func(x model.IndexedFullURL, _ int) string { return x.FullURL })

	for range maxStoreRetryCount {
		clear(full2shortBatch)

		for _, fullURL := range urlsToInsert {
			full2shortBatch[fullURL] = s.stringGenerator.GetRandomString(shortURLMaxLength)
		}

		retryToInsert, alreadyExists, storeErr := s.store.StoreBatch(ctx, full2shortBatch)

		traceLogger.Info().
			Strs("retryToInsert", retryToInsert).
			Strs("alreadyExists", lo.Keys(alreadyExists)).
			Send()

		if storeErr != nil {
			return nil, fmt.Errorf("error saving short url in store, %w", storeErr)
		}

		// if already added to result or need to retry, then don't add to result
		for k, v := range full2shortBatch {
			_, ready := readyBatch[k]
			if ready || lo.Contains(retryToInsert, k) {
				continue
			}

			if shortURL, storedOld := alreadyExists[k]; storedOld {
				readyBatch[k] = storeInfo{shortURL: shortURL, new: false}
			} else {
				readyBatch[k] = storeInfo{shortURL: v, new: true}
			}
		}

		// If no need retry, break loop
		if len(retryToInsert) == 0 {
			urlsToInsert = nil
			break
		}

		urlsToInsert = retryToInsert
	}

	result = lo.Map(request, func(x model.IndexedFullURL, _ int) model.IndexedShortURL {
		storedInfo := readyBatch[x.FullURL]
		link, _ := url.JoinPath(s.redirectAddr.String(), storedInfo.shortURL)
		return model.IndexedShortURL{CorrelationID: x.CorrelationID, ShortURL: link, IsCreated: storedInfo.new}
	})

	if len(urlsToInsert) != 0 {
		err = fmt.Errorf("no free space in storage, %w", err)
	}

	return
}
