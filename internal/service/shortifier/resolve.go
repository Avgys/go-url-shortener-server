package shortifier

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"go-url-shortener/internal/model"
	"go-url-shortener/internal/repository"
	httpShared "go-url-shortener/internal/shared/http"
)

func (s *Shortifier) ResolveShortURL(ctx context.Context, inputURL string, traceLogger *zerolog.Logger) (*model.DBURL, error) {

	shortURL := strings.TrimSpace(inputURL)

	url, err := s.store.ResolveShortURL(ctx, shortURL)

	if err != nil {

		if errors.Is(err, repository.ErrNotFound) {
			traceLogger.Info().
				Str("repository error", err.Error()).
				Msg("url not found in repository")

			return &url, httpShared.NewError("url not found", http.StatusNotFound)
		}

		traceLogger.Err(err).
			Send()

		return nil, err
	}

	if url.DeletedAtUTC != nil {
		return nil, httpShared.NewError("link deleted", http.StatusGone)
	}

	return &url, err
}
