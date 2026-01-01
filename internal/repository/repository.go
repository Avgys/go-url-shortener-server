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
	StoreNotFoundErr = errors.New("Url not found in store")
)

func (s *Store) StoreUrl(url string, urlHash string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, existed := s.data[urlHash]; existed {
		return false
	}

	s.data[urlHash] = url

	return true
}

func (s *Store) ResolveShortUrl(shortUrl string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	url := s.data[shortUrl]

	if url == "" {
		return "", StoreNotFoundErr
	}

	return url, nil
}
