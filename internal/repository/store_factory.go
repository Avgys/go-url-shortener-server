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

func NewRepository(done context.Context, cfg *config.Config, logger *zerolog.Logger) (Repository, error) {

	if cfg.DBConnectionString != "" {

		//Db
		dbConnection, err := db.NewDB(done, &db.Config{ConnectionString: cfg.DBConnectionString})

		if err != nil {
			return nil, err
		}

		return NewURLRepository(done, dbConnection, logger)
	}

	if cfg.FileStoragePath != "" {
		return NewFileStore(done, cfg.FileStoragePath, logger)
	}

	return NewInMemoryStore(nil), nil
}
