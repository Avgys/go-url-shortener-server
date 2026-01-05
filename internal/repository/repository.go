package repository

import (
	"sync"
)

type Store struct {
	Data map[string]string
	Mux  *sync.Mutex
}

func NewStore() *Store {
	return &Store{
		Data: make(map[string]string),
		Mux:  &sync.Mutex{},
	}
}

func (s *Store) StoreURL(url string, urlHash string) bool {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	if _, existed := s.Data[urlHash]; existed {
		return false
	}

	s.Data[urlHash] = url

	return true
}

func (s *Store) ResolveShortURL(shortURL string) (string, bool) {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	url, ok := s.Data[shortURL]

	return url, ok
}
