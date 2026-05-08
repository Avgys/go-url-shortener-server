package shortifier

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	db_model "go-url-shortener/internal/model/db"
	"go-url-shortener/internal/model/responses"
	"go-url-shortener/internal/repository"
	httpShared "go-url-shortener/internal/shared/http"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

func (s *Shortifier) GetURLsByUserID(ctx context.Context, userID int64, traceLogger *zerolog.Logger) ([]responses.URLPair, error) {

	dbURLs, err := s.store.GetURLsByUserID(ctx, userID)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		traceLogger.Info().
			Str("repository error", err.Error()).
			Msg("urls not found in repository")

		return nil, httpShared.NewError("no urls", http.StatusNoContent)
	}

	dbURLs = lo.Filter(dbURLs, func(h db_model.DBURL, _ int) bool { return h.DeletedAtUTC == nil })

	urls := lo.Map(dbURLs, func(dbURL db_model.DBURL, _ int) responses.URLPair {
		link, _ := url.JoinPath(s.redirectAddr.String(), dbURL.ShortURL)
		return responses.URLPair{ShortURL: link, OriginalURL: dbURL.OriginalURL}
	})

	return urls, err
}
