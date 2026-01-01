package repository

import (
	"errors"
	"sync"
)

type Store struct {
	data map[string]string
	mux  *sync.Mutex
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
		mux:  &sync.Mutex{},
	}
}

var (
	ErrStoreNotFound = errors.New("url not found in store")
)

func (s *Store) StoreURL(url string, urlHash string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, existed := s.data[urlHash]; existed {
		return false
	}

	s.data[urlHash] = url

	return true
}

func (s *Store) ResolveShortURL(shortURL string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	url := s.data[shortURL]

	if url == "" {
		return "", ErrStoreNotFound
	}

	return url, nil
}
