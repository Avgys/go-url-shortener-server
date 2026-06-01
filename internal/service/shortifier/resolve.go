package shortifier

import (
	"context"
	"errors"
	"net/http"
	"strings"

	dbmodel "go-url-shortener/internal/model/db"
	repoerrors "go-url-shortener/internal/repository/errors"
	httpShared "go-url-shortener/internal/shared/http"

	"github.com/rs/zerolog"
)

// ResolveShortURL looks up a short key in the store and returns the stored row.
// Missing URLs yield 404; soft-deleted URLs yield 410 Gone.
func (s *Shortifier) ResolveShortURL(ctx context.Context, inputURL string, traceLogger *zerolog.Logger) (*dbmodel.DBURL, error) {

	shortURL := strings.TrimSpace(inputURL)

	url, err := s.store.ResolveShortURL(ctx, shortURL)

	if err != nil {

		if errors.Is(err, repoerrors.ErrNotFound) {
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
