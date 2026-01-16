package repository

import (
	"errors"
	"maps"
	"sync"
)

var (
	ErrCollision = errors.New("slot in dictionary taken")
	ErrNotFound  = errors.New("url not found in store")
)

type Store struct {
	data map[string]string
	mux  sync.Mutex
}

func NewStore(initData map[string]string) *Store {
	s := &Store{data: make(map[string]string)}

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
