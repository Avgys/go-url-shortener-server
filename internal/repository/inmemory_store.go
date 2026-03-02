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
	storage storage
	mux     sync.Mutex
}

func NewInMemoryStore(initData []*model.DBURL) *InMemoryStore {

	mappedURLs := lo.Associate(initData, func(item *model.DBURL) (string, *model.DBURL) { return item.ShortURL, item })
	s := &InMemoryStore{storage: mappedURLs}

	return s
}

func (s *InMemoryStore) StoreBatch(ctx context.Context, input Full2ShortBatch, userID int64) (retryToInsert []string, alreadyExists map[string]string, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	alreadyExists = make(map[string]string)
	retryToInsert = make([]string, 0)

	maxURLID := lo.MaxBy(lo.Values(s.storage), func(a *model.DBURL, b *model.DBURL) bool { return a.ID > b.ID })

	var maxID = 0

	if maxURLID != nil {
		maxID = maxURLID.ID
	}

	for originURL, shortURL := range input {

		if storedValue, isStored := s.storage[shortURL]; isStored && storedValue.OriginalURL != originURL {
			retryToInsert = append(retryToInsert, originURL)
			continue
		}

		if value, ok := lo.FindKeyBy(s.storage, func(_ string, value *model.DBURL) bool { return value.OriginalURL == originURL }); ok {
			alreadyExists[originURL] = value
			continue
		}

		s.storage[shortURL] = &model.DBURL{
			ShortURL:    shortURL,
			OriginalURL: originURL,
			UserID:      userID,
			CreatedAt:   time.Now().UTC(),
			ID:          maxID,
		}

		maxID++
	}

	return
}

func (s *InMemoryStore) ResolveShortURL(ctx context.Context, shortURL string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	url, ok := s.storage[shortURL]

	if !ok {
		return "", ErrNotFound
	}

	return url.OriginalURL, nil
}

func (s *InMemoryStore) TestConnection(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) Close() error {
	return nil
}

func (s *InMemoryStore) getAll() []*model.DBURL {
	values := lo.Values(s.storage)
	return values
}

func (s *InMemoryStore) GetURLsByUserId(ctx context.Context, userID int64) ([]*model.DBURL, error) {

	values := lo.Values(s.storage)
	values = lo.Filter(values, func(item *model.DBURL, _ int) bool { return item.UserID == userID })

	return values, nil
}
