package repository

import (
	"context"
	"sync"
	"time"

	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/samber/lo"
)

type storage map[string]*model.DBURL

type InMemoryStore struct {
	shortURLToModel storage
	mux             sync.RWMutex
}

func NewInMemoryStore(initData []*model.DBURL) *InMemoryStore {

	mappedURLs := lo.Associate(initData, func(item *model.DBURL) (string, *model.DBURL) { return item.ShortURL, item })
	s := &InMemoryStore{shortURLToModel: mappedURLs}

	return s
}

func (s *InMemoryStore) StoreBatch(ctx context.Context, input Full2ShortBatch, userID int64) (retryToInsert []string, alreadyExists map[string]string, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	alreadyExists = make(map[string]string)
	retryToInsert = make([]string, 0)

	maxURLID := lo.MaxBy(lo.Values(s.shortURLToModel), func(a *model.DBURL, b *model.DBURL) bool { return a.ID > b.ID })

	var maxID = 0

	if maxURLID != nil {
		maxID = maxURLID.ID + 1
	}

	for originURL, shortURL := range input {

		if storedValue, isStored := s.shortURLToModel[shortURL]; isStored && storedValue.OriginalURL != originURL {
			retryToInsert = append(retryToInsert, originURL)
			continue
		}

		if value, ok := lo.FindKeyBy(s.shortURLToModel, func(_ string, value *model.DBURL) bool { return value.OriginalURL == originURL }); ok {
			alreadyExists[originURL] = value
			continue
		}

		s.shortURLToModel[shortURL] = &model.DBURL{
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

func (s *InMemoryStore) ResolveShortURL(ctx context.Context, shortURL string) (*model.DBURL, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	url, ok := s.shortURLToModel[shortURL]

	if !ok {
		return nil, ErrNotFound
	}

	return url, nil
}

func (s *InMemoryStore) TestConnection(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) Close() error {
	return nil
}

func (s *InMemoryStore) getAll() []*model.DBURL {
	defer s.mux.RUnlock()
	s.mux.RLock()

	values := lo.Values(s.shortURLToModel)
	return values
}

func (s *InMemoryStore) GetURLsByUserID(ctx context.Context, userID int64) ([]*model.DBURL, error) {

	s.mux.RLock()
	defer s.mux.RUnlock()
	result := make([]*model.DBURL, 0)

	for _, v := range s.shortURLToModel {

		if v.UserID == userID {

			result = append(result, v)
		}
	}

	return result, nil
}

func (s *InMemoryStore) DeleteURLS(context context.Context, groupedByUser map[int64][]string) ([]*model.DBURL, error) {

	s.mux.Lock()
	defer s.mux.Unlock()

	utcNow := time.Now().UTC()

	recordUpdate := make([]*model.DBURL, 0)

	for userID, urls := range groupedByUser {
		for _, url := range urls {
			dbURL, ok := s.shortURLToModel[url]

			if !ok {
				continue
			}

			if dbURL.UserID == userID {
				dbURL.DeletedAtUTC = &utcNow

				recordUpdate = append(recordUpdate, dbURL)
			}
		}
	}

	return recordUpdate, nil
}
