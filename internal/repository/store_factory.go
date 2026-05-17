package repository

import (
	"context"

	"go-url-shortener/internal/config"
	"go-url-shortener/internal/db"
	dbmodel "go-url-shortener/internal/model/db"
	"go-url-shortener/internal/repository/dbrepos"

	"github.com/rs/zerolog"
)

type Repository interface {
	StoreBatch(ctx context.Context, input map[string]string, userID int64) (retryToInsert []string, alreadyExists map[string]string, err error)
	ResolveShortURL(ctx context.Context, shortURL string) (dbmodel.DBURL, error)
	GetURLsByUserID(ctx context.Context, userID int64) ([]dbmodel.DBURL, error)
	TestConnection(ctx context.Context) error
	DeleteURLS(context context.Context, groupedByUser map[int64][]string) ([]dbmodel.DBURL, error)
	Close() error
}

func NewRepository(done context.Context, cfg *config.Config, logger *zerolog.Logger) (Repository, error) {

	if cfg.DBConnectionString != "" {

		//Db
		dbConnection, err := db.NewDB(done, &db.Config{ConnectionString: cfg.DBConnectionString})

		if err != nil {
			return nil, err
		}

		return dbrepos.NewURLRepository(done, dbConnection, logger), nil
	}

	if cfg.FileStoragePath != "" {
		return NewFileStore(done, cfg.FileStoragePath, logger)
	}

	return NewInMemoryStore(nil), nil
}
