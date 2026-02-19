package repository

import (
	"context"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/repository/db"
)

type Full2ShortBatch map[string]string

type Repository interface {
	StoreURL(ctx context.Context, url string, urlHash string) error
	StoreBatch(ctx context.Context, input Full2ShortBatch) (retryToInsert []string, alreadyExists map[string]string, err error)
	ResolveShortURL(ctx context.Context, shortURL string) (string, error)
	TestConnection(ctx context.Context) error
	Close() error
}

func NewRepository(ctx context.Context, cfg *config.Config) (Repository, error) {

	if cfg.DBConnectionString != "" {
		return NewDBStore(ctx, &db.Config{ConnectionString: cfg.DBConnectionString})
	}

	if cfg.FileStoragePath != "" {
		return NewFileStore(cfg.FileStoragePath)
	}

	return NewInMemoryStore(nil), nil
}
