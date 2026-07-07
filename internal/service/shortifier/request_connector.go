package shortifier

import (
	"context"
	"errors"
	"fmt"
	"go-url-shortener/internal/model/requests"
	"go-url-shortener/internal/model/responses"
	"go-url-shortener/internal/service/auth"
	shared "go-url-shortener/internal/shared/http"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
)

// ShortenURL shortens a single URL using the authenticated user from r's context.
// It publishes an audit event for each stored long URL.
func (s *Shortifier) ShortenURL(ctx context.Context, url string, traceLogger *zerolog.Logger) (*responses.IndexedShortURL, error) {
	batch := requests.ShortenBatchReq{requests.IndexedFullURL{FullURL: url}}
	urls, err := s.ShortenBatch(ctx, batch, traceLogger)

	if err != nil {
		return nil, err
	}

	return &(*urls)[0], nil
}

// ShortenBatch shortens multiple URLs using the authenticated user from r's context.
// It publishes an audit event per input URL after a successful store.
func (s *Shortifier) ShortenBatch(ctx context.Context, batch requests.ShortenBatchReq, traceLogger *zerolog.Logger) (*responses.ShortenBatchResp, error) {

	claims, err := auth.GetFromContext(ctx)

	if err != nil {
		showErr := shared.NewError("unauthorized/broken token", http.StatusUnauthorized)
		return nil, fmt.Errorf("%w inner error: %w", showErr, err)
	}

	serviceReq := &ShortenBatchReq{URLs: batch, UserID: claims.UserID}
	resultURL, err := s.ShortifyBatch(ctx, serviceReq, traceLogger)

	if err != nil {
		if errors.Is(err, ErrCollision) {
			err = shared.NewError(err.Error(), http.StatusTooManyRequests)
		}

		return nil, err
	}

	for _, url := range batch {
		s.auditService.Publish("shorten", strconv.FormatInt(claims.UserID, 10), url.FullURL)
	}

	return &resultURL, nil
}
