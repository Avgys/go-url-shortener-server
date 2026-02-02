package repository

import "errors"

var (
	ErrCollision = errors.New("slot in dictionary taken")
	ErrNotFound  = errors.New("url not found in store")
)

type Repository interface {
	StoreURL(url string, urlHash string) error
	ResolveShortURL(shortURL string) (string, error)
	getAll() storage
}
