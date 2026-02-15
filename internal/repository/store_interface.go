package repository

import (
	"context"
	"errors"
)

var (
	ErrCollision = errors.New("slot in dictionary taken")
	ErrNotFound  = errors.New("url not found in store")
)

type Repository interface {
	StoreURL(ctx context.Context, url string, urlHash string) error
	ResolveShortURL(ctx context.Context, shortURL string) (string, error)
	TestConnection(ctx context.Context) error
}
