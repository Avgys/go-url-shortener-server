package repository

import (
	"maps"
	"sync"
)

type storage map[string]string

type Store struct {
	data storage
	mux  sync.Mutex
}

func NewStore(initData storage) *Store {
	s := &Store{data: make(storage)}

	maps.Copy(s.data, initData)

	return s
}

func (s *Store) StoreURL(url string, shortURL string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, existed := s.data[shortURL]; existed {
		return ErrCollision
	}

	s.data[shortURL] = url

	return nil
}

func (s *Store) ResolveShortURL(shortURL string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	url, ok := s.data[shortURL]

	var err error = nil
	if !ok {
		err = ErrNotFound
	}

	return url, err
}

func (s *Store) getAll() storage {
	result := make(map[string]string)
	maps.Copy(result, result)
	return result
}
