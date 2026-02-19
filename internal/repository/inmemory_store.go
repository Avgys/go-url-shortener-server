package repository

import (
	"context"
	"maps"
	"sync"

	"github.com/samber/lo"
)

type storage map[string]string

type InMemoryStore struct {
	short2long storage
	mux        sync.Mutex
}

func NewInMemoryStore(initData storage) *InMemoryStore {
	s := &InMemoryStore{short2long: make(storage)}

	maps.Copy(s.short2long, initData)

	return s
}

func (s *InMemoryStore) StoreBatch(ctx context.Context, input Full2ShortBatch) (retryToInsert []string, alreadyExists map[string]string, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	alreadyExists = make(map[string]string)
	retryToInsert = make([]string, 0)

	for longURL, shortURL := range input {

		if storedValue, isStored := s.short2long[shortURL]; isStored && storedValue != longURL {
			retryToInsert = append(retryToInsert, longURL)
			continue
		}

		if value, ok := lo.FindKey(s.short2long, longURL); ok {
			alreadyExists[longURL] = value
			continue
		}

		s.short2long[shortURL] = longURL
	}

	return
}

func (s *InMemoryStore) ResolveShortURL(ctx context.Context, shortURL string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	url, ok := s.short2long[shortURL]

	var err error = nil
	if !ok {
		err = ErrNotFound
	}

	return url, err
}

func (s *InMemoryStore) TestConnection(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) Close() error {
	return nil
}

func (s *InMemoryStore) getAll() *storage {
	result := make(map[string]string)
	maps.Copy(result, s.short2long)
	store := storage(result)

	return &store
}
