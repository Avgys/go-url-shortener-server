package repository

import (
	"context"
	"maps"
	"sync"
)

type storage map[string]string

type InMemoryStore struct {
	data storage
	mux  sync.Mutex
}

func NewInMemoryStore(initData storage) *InMemoryStore {
	s := &InMemoryStore{data: make(storage)}

	maps.Copy(s.data, initData)

	return s
}

func (s *InMemoryStore) StoreURL(ctx context.Context, url string, shortURL string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, existed := s.data[shortURL]; existed {
		return ErrCollision
	}

	s.data[shortURL] = url

	return nil
}

func (s *InMemoryStore) ResolveShortURL(ctx context.Context, shortURL string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	url, ok := s.data[shortURL]

	var err error = nil
	if !ok {
		err = ErrNotFound
	}

	return url, err
}

func (s *InMemoryStore) TestConnection(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) getAll() *storage {
	result := make(map[string]string)
	maps.Copy(result, s.data)
	store := storage(result)

	return &store
}
