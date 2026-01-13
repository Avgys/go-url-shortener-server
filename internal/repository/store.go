package repository

import (
	"errors"
	"sync"
)

var (
	ErrCollision = errors.New("slot in dictionary taken")
	ErrNotFound  = errors.New("url not found in store")
)

type Store struct {
	Data map[string]string
	mux  sync.Mutex
}

func NewStore(initData map[string]string) *Store {
	s := &Store{Data: make(map[string]string)}

	for k, v := range initData {
		s.Data[k] = v
	}

	return s
}

func (s *Store) StoreURL(url string, shortURL string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, existed := s.Data[shortURL]; existed {
		return false
	}

	s.Data[shortURL] = url

	return true
}

func (s *Store) ResolveShortURL(shortURL string) (string, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	url, ok := s.Data[shortURL]

	return url, ok
}
