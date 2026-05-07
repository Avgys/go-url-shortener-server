package repository

import (
	"context"

	"go-url-shortener/internal/config"
	"go-url-shortener/internal/db"
	"go-url-shortener/internal/model"

	"github.com/rs/zerolog"
)

type Full2ShortBatch map[string]string

type Repository interface {
	StoreBatch(ctx context.Context, input Full2ShortBatch, userID int64) (retryToInsert []string, alreadyExists map[string]string, err error)
	ResolveShortURL(ctx context.Context, shortURL string) (model.DBURL, error)
	GetURLsByUserID(ctx context.Context, userID int64) ([]model.DBURL, error)
	TestConnection(ctx context.Context) error
	DeleteURLS(context context.Context, groupedByUser map[int64][]string) ([]model.DBURL, error)
	Close() error
}

func NewRepository(ctx context.Context, cfg *config.Config, logger *zerolog.Logger) (Repository, error) {

	if cfg.DBConnectionString != "" {
		return NewDBStore(ctx, &db.Config{ConnectionString: cfg.DBConnectionString}, logger)
	}

	if cfg.FileStoragePath != "" {
		return NewFileStore(ctx, cfg.FileStoragePath, logger)
	}

	return NewInMemoryStore(nil), nil
}
