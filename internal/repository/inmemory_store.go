package repository

import (
	"context"
	"sync"
	"time"

	dbmodel "go-url-shortener/internal/model/db"
	repoerrors "go-url-shortener/internal/repository/errors"

	"github.com/samber/lo"
)

type storage map[string]*dbmodel.DBURL

type InMemoryStore struct {
	shortURLToModel storage
	mux             sync.RWMutex
}

func NewInMemoryStore(initData []*dbmodel.DBURL) *InMemoryStore {

	mappedURLs := lo.Associate(initData, func(item *dbmodel.DBURL) (string, *dbmodel.DBURL) { return item.ShortURL, item })
	s := &InMemoryStore{shortURLToModel: mappedURLs}

	return s
}

func (s *InMemoryStore) StoreBatch(ctx context.Context, input map[string]string, userID int64) (retryToInsert []string, alreadyExists map[string]string, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	alreadyExists = make(map[string]string)
	retryToInsert = make([]string, 0)

	maxURLID := lo.MaxBy(lo.Values(s.shortURLToModel), func(a *dbmodel.DBURL, b *dbmodel.DBURL) bool { return a.ID > b.ID })

	var maxID = 0

	if maxURLID != nil {
		maxID = maxURLID.ID + 1
	}

	for originURL, shortURL := range input {

		if storedValue, isStored := s.shortURLToModel[shortURL]; isStored && storedValue.OriginalURL != originURL {
			retryToInsert = append(retryToInsert, originURL)
			continue
		}

		if value, ok := lo.FindKeyBy(s.shortURLToModel, func(_ string, value *dbmodel.DBURL) bool { return value.OriginalURL == originURL }); ok {
			alreadyExists[originURL] = value
			continue
		}

		s.shortURLToModel[shortURL] = &dbmodel.DBURL{
			ShortURL:     shortURL,
			OriginalURL:  originURL,
			UserID:       userID,
			CreatedAt:    time.Now().UTC(),
			ID:           maxID,
			DeletedAtUTC: nil,
		}

		maxID++
	}

	return
}

func (s *InMemoryStore) ResolveShortURL(ctx context.Context, shortURL string) (dbmodel.DBURL, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	url, ok := s.shortURLToModel[shortURL]

	if !ok {
		return dbmodel.DBURL{}, repoerrors.ErrNotFound
	}

	return *url, nil
}

func (s *InMemoryStore) TestConnection(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) Close() error {
	return nil
}

func (s *InMemoryStore) getAll() []dbmodel.DBURL {
	defer s.mux.RUnlock()
	s.mux.RLock()

	values := lo.Values(s.shortURLToModel)
	return lo.Map(values, func(item *dbmodel.DBURL, _ int) dbmodel.DBURL { return *item })
}

func (s *InMemoryStore) GetURLsByUserID(ctx context.Context, userID int64) ([]dbmodel.DBURL, error) {

	s.mux.RLock()
	defer s.mux.RUnlock()
	result := make([]dbmodel.DBURL, 0)

	for _, v := range s.shortURLToModel {

		if v.UserID == userID {

			result = append(result, *v)
		}
	}

	return result, nil
}

func (s *InMemoryStore) DeleteURLS(context context.Context, groupedByUser map[int64][]string) ([]dbmodel.DBURL, error) {

	s.mux.Lock()
	defer s.mux.Unlock()

	utcNow := time.Now().UTC()

	recordUpdate := make([]dbmodel.DBURL, 0)

	for userID, urls := range groupedByUser {
		for _, url := range urls {
			dbURL, ok := s.shortURLToModel[url]

			if !ok {
				continue
			}

			if dbURL.UserID == userID {
				dbURL.DeletedAtUTC = &utcNow

				recordUpdate = append(recordUpdate, *dbURL)
			}
		}
	}

	return recordUpdate, nil
}

func (s *InMemoryStore) GetURLsStats(ctx context.Context) (dbmodel.GetURLsStatsRow, error) {
	_ = ctx

	s.mux.RLock()
	defer s.mux.RUnlock()

	uniqueLongURLs := make(map[string]struct{}, len(s.shortURLToModel))
	uniqueUserIDs := make(map[int64]struct{}, len(s.shortURLToModel))

	for _, row := range s.shortURLToModel {
		if row == nil {
			continue
		}

		uniqueLongURLs[row.OriginalURL] = struct{}{}
		uniqueUserIDs[row.UserID] = struct{}{}
	}

	return dbmodel.GetURLsStatsRow{
		UniqueLongUrlCount: int64(len(uniqueLongURLs)),
		UniqueUserIDCount:  int64(len(uniqueUserIDs)),
	}, nil
}
